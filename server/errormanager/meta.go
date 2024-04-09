package errormanager

import (
	"context"
	"os"

	"gitee.com/openeuler/PilotGo/sdk/logger"
)

type Topoerror struct {
	Err    error
	Cancel context.CancelFunc
}

/*
@ctx:	插件客户端初始上下文（默认为agentmanager.Topo.Tctx）

@err:	最终生成的error

@exit_after_print: 打印完错误链信息后是否结束主程序
*/
func ErrorTransmit(ctx context.Context, err error, exit_after_print bool) {
	if GlobalErrorManager == nil {
		logger.Error("globalerrormanager is nil")
		os.Exit(1)
	}

	if exit_after_print {
		cctx, cancelF := context.WithCancel(ctx)
		GlobalErrorManager.ErrCh <- &Topoerror{
			Err:    err,
			Cancel: cancelF,
		}
		<-cctx.Done()
		close(GlobalErrorManager.ErrCh)
		os.Exit(1)
	}

	GlobalErrorManager.ErrCh <- &Topoerror{
		Err: err,
	}
}
