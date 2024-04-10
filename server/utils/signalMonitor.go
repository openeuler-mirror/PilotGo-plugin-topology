package utils

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func SignalMonitoring(neo4jclient neo4j.Driver, redisclient redis.Client) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for s := range ch {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			neo4jclient.Close()
			fmt.Println()
			logger.Info("close the connection to neo4j\n")
			redisclient.Close()
			logger.Info("close the connection to redis\n")
			os.Exit(1)
		default:
			logger.Warn("unknown signal-> %s\n", s.String())
		}
	}
}
