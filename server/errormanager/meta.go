package errormanager

import (
	"context"
	"os"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

type Topoerror struct {
	Err    error
	Cancel context.CancelFunc
}

/*
@ctx:	插件服务端初始上下文（默认为pluginclient.Global_Context）

@err:	最终生成的error

@exit_after_print: 打印完错误链信息后是否结束主程序
*/
func ErrorTransmit(_ctx context.Context, _err error, _exit_after_print bool) {
	if Global_ErrorManager == nil {
		logger.Error("globalerrormanager is nil")
		global.Close()
		os.Exit(1)
	}

	if _exit_after_print {
		cctx, cancelF := context.WithCancel(_ctx)
		Global_ErrorManager.ErrCh <- &Topoerror{
			Err:    _err,
			Cancel: cancelF,
		}
		<-cctx.Done()
		close(Global_ErrorManager.ErrCh)
		global.Close()
		os.Exit(1)
	}

	Global_ErrorManager.ErrCh <- &Topoerror{
		Err: _err,
	}
}
