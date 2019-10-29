#!/bin/bash
echo ">>>>>>>>>>>>>>>>>>>>>>>>> 正在打 Linux 环境包 >>>>>>>>>>>>>>>>>>>>>>>>>"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pkg/linux/flight_go_linux .
echo ">>>>>>>>>>>>>>>>>>>>>>>>> Linux 环境包打包完成 >>>>>>>>>>>>>>>>>>>>>>>>>"

echo ">>>>>>>>>>>>>>>>>>>>>>>>> 正在打 Windows 环境包 >>>>>>>>>>>>>>>>>>>>>>>>>"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o pkg/win64/flight_go_win64.exe .
echo ">>>>>>>>>>>>>>>>>>>>>>>>> Windows 环境包打包完成 >>>>>>>>>>>>>>>>>>>>>>>>>"

echo ">>>>>>>>>>>>>>>>>>>>>>>>> 正在打 Mac 环境包 >>>>>>>>>>>>>>>>>>>>>>>>>"
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o pkg/darwin/flight_go_darwin .
echo ">>>>>>>>>>>>>>>>>>>>>>>>> Mac 环境包打包完成 >>>>>>>>>>>>>>>>>>>>>>>>>"