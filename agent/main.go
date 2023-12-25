package main

import (
	"fmt"
	"os"

	"gitee.com/openeuler/PilotGo-plugin-topology-agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-agent/handler"
	"gitee.com/openeuler/PilotGo-plugin-topology-agent/service"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/gin-gonic/gin"
)

func main() {

	InitLogger()

	go func() {
		engine := gin.Default()
		handler.InitRouter(engine)
		if err := engine.Run(conf.Config().Topo.Agent_addr); err != nil {
			logger.Fatal("failed to run web server")
		}
	}()

	service.SendHeartbeat()
}

func InitLogger() {
	if err := logger.Init(conf.Config().Logopts); err != nil {
		fmt.Printf("logger init failed, please check the config file: %s", err)
		os.Exit(1)
	}
}
