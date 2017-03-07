package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/text/encoding/japanese"
)

var version string

func processLinesShiftJIS(lineProcessor func(string), r io.Reader, wg *sync.WaitGroup) {
	decoder := japanese.ShiftJIS.NewDecoder()
	scanner := bufio.NewScanner(decoder.Reader(r))
	for scanner.Scan() {
		line := scanner.Text()
		lineProcessor(line)
	}
	wg.Done()
}

func outputStdout(line string) {
	if line != "" {
		log.Println("o:", line)
	}
}

func outputStderr(line string) {
	if line != "" {
		log.Println("e:", line)
	}
}

func runCommand(stdoutHandler, stderrHandler func(string), name string, arg ...string) (err error) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("StdoutPipe:", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println("StderrPipe:", err)
		return
	}

	err = cmd.Start()
	if err != nil {
		log.Println("Start:", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go processLinesShiftJIS(stdoutHandler, stdout, &wg)
	go processLinesShiftJIS(stderrHandler, stderr, &wg)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Println("Wait:", err)
		return
	}
	return nil
}

var spacesRe = regexp.MustCompile(`\s+`)

func (mw *MyMainWindow) getInterfaces() {
	log.Printf("インターフェイス一覧を取得します。")
	interfaces := []string{}
	index, select_index := 0, 0
	err := runCommand(func(line string) {
		outputStdout(line)
		a := spacesRe.Split(strings.TrimSpace(line), 5)
		if len(a) == 5 && (a[3] == "connected" || a[3] == "disconnected") {
			if strings.Contains(a[4], "Loopback") {
				return
			}
			interfaces = append(interfaces, a[4])
			if strings.Contains(a[4], "ローカル エリア接続") {
				select_index = index
			}
			index += 1
		}
	}, outputStderr, "netsh", "interface", "ip", "show", "interfaces")
	if err != nil {
		log.Printf("インターフェイス一覧の取得に失敗しました。")
		return
	}
	mw.interfaces = interfaces
	mw.lb.SetModel(mw.interfaces)
	mw.lb.SetCurrentIndex(select_index)
	log.Printf("インターフェイス一覧を取得しました。%q", interfaces)
}

func setDhcp(iface string) {
	log.Printf("%sをDHCPに設定します。", iface)
	err := runCommand(outputStdout, outputStderr, "netsh", "interface", "ip", "set", "address", iface, "dhcp")
	if err != nil {
		log.Printf("%sをDHCPに設定できませんでした。", iface)
		return
	}
	log.Printf("%sをDHCPに設定しました。", iface)
}

func setStatic(iface, ip string, callback func(string)) {
	log.Printf("%sを%sに設定します。", iface, ip)
	err := runCommand(outputStdout, outputStderr, "netsh", "interface", "ip", "set", "address", iface, "static", ip, "255.255.255.0")
	if err != nil {
		log.Printf("%sを固定IPに設定できませんでした。", iface)
		return
	}
	log.Printf("%sを%sに設定しました。", iface, ip)
	callback(ip)
}

type MyMainWindow struct {
	*walk.MainWindow
	lb *walk.ListBox
	cb *walk.ComboBox

	interfaces []string
}

func (mw *MyMainWindow) getInterface() (string, error) {
	idx := mw.lb.CurrentIndex()
	if idx < 0 {
		return "", errors.New("インターフェイスを選択してください。")
	}
	return mw.interfaces[mw.lb.CurrentIndex()], nil
}

func (mw *MyMainWindow) appendIp(ip string) {
	ip_addresses, ok := mw.cb.Model().([]string)
	if !ok {
		walk.MsgBox(mw, "内部エラー", "ip_addresses 取得エラー", walk.MsgBoxIconError)
		return
	}
	for _, x := range ip_addresses {
		if x == ip {
			return
		}
	}
	mw.cb.SetModel(append(ip_addresses, ip))
	mw.cb.SetCurrentIndex(len(ip_addresses))
}

func (mw *MyMainWindow) setStatic() {
	iface, err := mw.getInterface()
	if err != nil {
		walk.MsgBox(mw, "エラー", fmt.Sprint(err), walk.MsgBoxIconError)
		return
	}
	go setStatic(iface, mw.cb.Text(), mw.appendIp)
}

func (mw *MyMainWindow) setDhcp() {
	iface, err := mw.getInterface()
	if err != nil {
		walk.MsgBox(mw, "エラー", fmt.Sprint(err), walk.MsgBoxIconError)
		return
	}
	go setDhcp(iface)
}

func main() {
	mw := &MyMainWindow{}
	ip_addresses := []string{
		"192.168.0.1",
		"192.168.1.2",
		"192.168.3.2",
		"192.168.10.2",
	}

	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "IP設定",
		MinSize:  Size{600, 400},
		Layout:   VBox{},
		Children: []Widget{
			ListBox{
				AssignTo: &mw.lb,
				Model:    mw.interfaces,
			},
			ComboBox{
				AssignTo: &mw.cb,
				Editable: true,
				Model:    ip_addresses,
				OnKeyPress: func(k walk.Key) {
					if k == walk.KeyReturn {
						mw.setStatic()
					}
				},
			},
			PushButton{
				Text:      "固定IP設定",
				OnClicked: mw.setStatic,
			},
			PushButton{
				Text:      "DHCP設定",
				OnClicked: mw.setDhcp,
			},
			PushButton{
				Text: "インターフェイス一覧再取得",
				OnClicked: func() {
					go mw.getInterfaces()
				},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	lv, err := NewLogView(mw)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(lv)
	log.Println("winipset バージョン", version)

	go mw.getInterfaces()

	mw.Run()
}
