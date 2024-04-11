package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/graphmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/redismanager"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func SignalMonitoring() {
	// main.go中最后执行，不进行全局指针变量非空判断
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for s := range ch {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			switch conf.Global_Config.Topo.GraphDB {
			case "neo4j":
				graphmanager.Global_Neo4j.Driver.Close()
				fmt.Println()
			}

			logger.Info("close the connection to neo4j\n")
			redismanager.Global_Redis.Client.Close()
			logger.Info("close the connection to redis\n")
			os.Exit(1)
		default:
			logger.Warn("unknown signal-> %s\n", s.String())
		}
	}
}
