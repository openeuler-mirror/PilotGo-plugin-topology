package global

import (
	"context"
	"fmt"
	"sync"

	"gitee.com/openeuler/PilotGo/sdk/logger"
)

var END *ResourceRelease

func init() {
	ctx, cancel := context.WithCancel(RootContext)
	END = &ResourceRelease{
		CancelCtx:  ctx,
		CancelFunc: cancel,
	}
}

type ResourceRelease struct {
	Wg sync.WaitGroup

	CancelCtx  context.Context
	CancelFunc context.CancelFunc
}

func (rr *ResourceRelease) Close() {
	switch Global_graph_database {
	case "neo4j":
		if Global_neo4j_driver != nil {
			Global_neo4j_driver.Close()
			fmt.Println()
			logger.Info("close the connection to neo4j\n")
		}
	}

	if Global_redis_client != nil {
		Global_redis_client.Close()
		logger.Info("close the connection to redis\n")
	}

	if Global_influx_client != nil {
		Global_influx_client.Close()
		logger.Info("close the connection to influx\n")
	}

	rr.CancelFunc()

	rr.Wg.Wait()
}
