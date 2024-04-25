package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/graphmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/influxmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/redismanager"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func SignalMonitoring() {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for s := range ch {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			switch conf.Global_Config.Topo.GraphDB {
			case "neo4j":
				if graphmanager.Global_Neo4j != nil {
					graphmanager.Global_Neo4j.Driver.Close()
					fmt.Println()
					logger.Info("close the connection to neo4j\n")
				}
			}

			if redismanager.Global_Redis != nil {
				redismanager.Global_Redis.Client.Close()
				logger.Info("close the connection to redis\n")
			}

			if influxmanager.Global_Influx != nil {
				influxmanager.Global_Influx.Client.Close()
				logger.Info("close the connection to influx\n")
			}

			os.Exit(1)
		default:
			logger.Warn("unknown signal-> %s\n", s.String())
		}
	}
}
