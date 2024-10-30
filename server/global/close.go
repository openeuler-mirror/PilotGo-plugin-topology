package global

import (
	"fmt"

	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func Close() {
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

	Global_cancelFunc()

	Global_wg.Wait()
}
