# Windows IP 設定ラッパー

ネットワーク機器の設定の時に固定 IP アドレスの設定を簡略化するための netsh のラッパープログラムです。

## How to build

- `make`
- `make GOARCH=amd64`

## How to use

- ドメイン環境などで管理者権限が得られない場合は、ローカルの Network Configuration Operators のグループに所属すれば IP アドレスの変更ができます。

## LICENSE

- logview.go: BSD-style license (original is https://github.com/lxn/walk/blob/050d2729e78b39c5bc8335e744758910f2c7a6c2/examples/logview/logview.go )
- The MIT License (MIT)
