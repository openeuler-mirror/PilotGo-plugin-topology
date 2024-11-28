/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
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
