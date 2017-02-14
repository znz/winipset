# Windows IP 設定ラッパー

ネットワーク機器の設定の時に固定 IP アドレスの設定を簡略化するための netsh のラッパープログラムです。

## How to build

- `GOOS=windows go get github.com/lxn/walk`
- `GOOS=windows go get golang.org/x/text/encoding/japanese`
- `go get github.com/akavel/rsrc`
- `rsrc -manifest winipset.manifest -o rsrc.syso`
- `GOOS=windows go build -ldflags="-H windowsgui"`

## LICENSE

- logview.go: BSD-style license (original is https://github.com/lxn/walk/blob/050d2729e78b39c5bc8335e744758910f2c7a6c2/examples/logview/logview.go )
- The MIT License (MIT)
