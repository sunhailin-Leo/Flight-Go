package main

import (
	"os"
	"time"
)

const (
	serviceName           string = "Flight-Go"
	currentServiceVersion string = "v0.1.0"
	logMaxAge                    = time.Hour * 24
	logRotationTime              = time.Hour * 24
	logPath               string = ""
	logFileName           string = "Flight-Go.log"
)

func main() {
	// 初始化日志
	logger = NewLogger(logMaxAge, logRotationTime, logPath, logFileName)
	// 初始化数据
	initCityNameCodeData()
	// 命令行初始化
	commandLineInit()
	commands := flightCommands
	args := os.Args
	if len(args) > 1 {
		for _, cmd := range commands {
			if cmd.Run != nil && cmd.Name() == args[1] {
				err := cmd.Flag.Parse(args[2:])
				if err != nil {
					os.Exit(1)
				}
				args = cmd.Flag.Args()
				if len(args) > 0 {
					os.Exit(cmd.Run(args))
				}
				logger.Errorf("[Flight-Go]命令参数错误!")
				break
			}
		}
	} else {
		logger.Errorf("[Flight-Go]命令参数错误!")
	}
}
