package signal

import (
	"os"
	"os/signal"
	"syscall"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func SignalMonitoring() {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for s := range ch {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			global.ERManager.ResourceRelease()
			os.Exit(1)
		default:
			logger.Warn("unknown signal-> %s\n", s.String())
		}
	}
}
