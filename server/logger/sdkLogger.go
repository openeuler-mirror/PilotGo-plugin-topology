package logger

import (
	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
)

func InitLogger() {
	err := logger.Init(conf.Global_Config.Logopts)
	if err != nil {
		err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
	}
}
