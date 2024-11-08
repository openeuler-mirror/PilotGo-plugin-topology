package logger

import (
	"fmt"
	"os"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/conf"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func InitLogger() {
	err := logger.Init(conf.Global_Config.Logopts)
	if err != nil {
		fmt.Printf("logger module init failed: %s\n", err.Error())
		os.Exit(1)
	}
}
