package global

import (
	"errors"
	"fmt"
)

func Close() {
	switch Global_graph_database {
	case "neo4j":
		if Global_neo4j_driver != nil {
			Global_neo4j_driver.Close()
			fmt.Println()
			ERManager.ErrorTransmit("global", "info", errors.New("close the connection to neo4j"), false, false)
		}
	}

	if Global_redis_client != nil {
		Global_redis_client.Close()
		ERManager.ErrorTransmit("global", "info", errors.New("close the connection to redis"), false, false)
	}

	if Global_influx_client != nil {
		Global_influx_client.Close()
		ERManager.ErrorTransmit("global", "info", errors.New("close the connection to influx"), false, false)
	}
}
