package signal

import (
	"os"
	"os/signal"
	"syscall"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"github.com/pkg/errors"
)

func SignalMonitoring() {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for s := range ch {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			global.ERManager.ErrorTransmit("signal", "info", errors.Errorf("signal interrupt: %s", s.String()), false, false)
			global.ERManager.ResourceRelease()
			os.Exit(1)
		default:
			global.ERManager.ErrorTransmit("signal", "warn", errors.Errorf("unknown signal: %s", s.String()), false, false)
		}
	}
}
