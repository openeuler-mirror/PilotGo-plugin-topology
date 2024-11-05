package resourcemanage

import (
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func InitResourceManager() {
	ermanager, err := CreateErrorReleaseManager(global.RootContext, global.Close)
	if err != nil {
		logger.Fatal(err.Error())
	}
	ERManager = ermanager
}
