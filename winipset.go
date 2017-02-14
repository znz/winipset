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
	"syscall"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/text/encoding/japanese"
)

func getInterfaces(mw *MyMainWindow) {
	cmd := exec.Command("netsh", "interface", "ip", "show", "interfaces")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println(err)
		return
	}

	if err = cmd.Start(); err != nil {
		log.Println(err)
		return
	}

	go printOutput(stderr)

	decoder := japanese.ShiftJIS.NewDecoder()
	scanner := bufio.NewScanner(decoder.Reader(stdout))
	re := regexp.MustCompile(`\s+`)
	interfaces := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		a := re.Split(strings.TrimSpace(line), 5)
		if len(a) == 5 && (a[3] == "connected" || a[3] == "disconnected") {
			if strings.Contains(a[4], "Loopback") {
				continue
			}
			interfaces = append(interfaces, a[4])
		}
	}
	log.Printf("%q", interfaces)

	if err = cmd.Wait(); err != nil {
		log.Println(err)
		return
	}
	mw.interfaces = interfaces
	mw.lb.SetModel(mw.interfaces)
	mw.lb.SetCurrentIndex(0)
}

func printOutput(r io.Reader) {
	decoder := japanese.ShiftJIS.NewDecoder()
	scanner := bufio.NewScanner(decoder.Reader(r))
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			log.Print(line)
		}
	}
}

func runCommand(message, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println(err)
		return
	}

	if err = cmd.Start(); err != nil {
		log.Println(err)
		return
	}

	go printOutput(stdout)
	go printOutput(stderr)

	if err = cmd.Wait(); err != nil {
		log.Println(err)
		return
	}
	log.Println(message)
}

func setDhcp(iface string) {
	runCommand(fmt.Sprintf("%sをDHCPに設定しました。", iface), "netsh", "interface", "ip", "set", "address", iface, "dhcp")
}

func setStatic(iface, ip string) {
	runCommand(fmt.Sprintf("%sを%sに設定しました。", iface, ip), "netsh", "interface", "ip", "set", "address", iface, "static", ip, "255.255.255.0")
}

type MyMainWindow struct {
	*walk.MainWindow
	lb *walk.ListBox
	cb *walk.ComboBox

	interfaces []string
}

func getInterface(mw *MyMainWindow) (string, error) {
	idx := mw.lb.CurrentIndex()
	if idx < 0 {
		return "", errors.New("インターフェイスを選択してください。")
	}
	return mw.interfaces[mw.lb.CurrentIndex()], nil
}

func main() {
	mw := &MyMainWindow{}
	ip_addresses := []string{
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
			},
			PushButton{
				Text: "固定IP設定",
				OnClicked: func() {
					iface, err := getInterface(mw)
					if err != nil {
						walk.MsgBox(mw, "エラー", fmt.Sprint(err), walk.MsgBoxIconError)
						return
					}
					go setStatic(iface, mw.cb.Text())
				},
			},
			PushButton{
				Text: "DHCP設定",
				OnClicked: func() {
					iface, err := getInterface(mw)
					if err != nil {
						walk.MsgBox(mw, "エラー", fmt.Sprint(err), walk.MsgBoxIconError)
						return
					}
					go setDhcp(iface)
				},
			},
			PushButton{
				Text: "インターフェイス一覧再取得",
				OnClicked: func() {
					go getInterfaces(mw)
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

	go getInterfaces(mw)

	mw.Run()
}
