/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
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
