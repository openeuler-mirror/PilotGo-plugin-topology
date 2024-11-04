package main

import (
	"fmt"
	"os"
	"runtime"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/handler"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/service"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/signal"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func main() {
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

	InitLogger()

	handler.InitWebServer()

	service.SendHeartbeat()

	signal.SignalMonitoring()
}

func InitLogger() {
	if err := logger.Init(conf.Config().Logopts); err != nil {
		fmt.Printf("logger init failed, please check the config file: %s", err)
		os.Exit(1)
	}
}
