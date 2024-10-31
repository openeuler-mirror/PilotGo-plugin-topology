package logger

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func InitLogger() {
	err := logger.Init(conf.Global_Config.Logopts)
	if err != nil {
		fmt.Printf("logger module init failed: %s\n", err.Error())
	}
}
