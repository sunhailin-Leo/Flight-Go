<h1 align="center">Flight-Go</h1>
<p align="center">
    <em>Use Go-resty to crawl ctrip</em>
</p>
<p align="center">
    <a href="https://github.com/sunhailin-Leo">
        <img src="https://img.shields.io/badge/Author-sunhailin--Leo-blue" alt="Author">
    </a>
</p>
<p align="center">
    <a href="https://opensource.org/licenses/MIT">
        <img src="https://img.shields.io/badge/License-MIT-brightgreen.svg" alt="License">
    </a>
</p>

## 💯 项目说明

* 项目包管理基于 [govendor](https://github.com/kardianos/govendor) 构建，项目使用了 [go-resty](https://github.com/go-resty/resty) 作为 HTTP 请求框架
* 打包文件在 `pkg` 文件夹中（darwin 对应 Mac OS，linux 对应 Linux 系统，win64 对应 Windows 64位系统）

## 💻 使用说明

**Linux / Mac OS 下使用**
```shell script
chmod a+x flight_go
# 查询机票价格信息
./flight_go schedule <起飞机场> <到达机场> <当前日期(日期格式: YYYY-MM-DD)>
```

**Windows 下使用(Windows 控制台下)**
```shell script
# 查询机票价格信息
flight_go.exe schedule <起飞机场> <到达机场> <当前日期(日期格式: YYYY-MM-DD)>
```

## 📖 功能说明

* 目前暂时开发了两个功能:
    * Version v0.1.0
        * 通过日期查询两地航班信息

* 后续开发功能点:
    * 命令行参数提示
    * 加入代理配置
    * 争取完善一些命令行交互以及其他查询功能

## 📃 License

MIT [©sunhailin-Leo](https://github.com/sunhailin-Leo)