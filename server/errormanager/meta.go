package errormanager

import (
	"context"
	"os"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

type FinalError struct {
	Err error

	Severity string

	Cancel context.CancelFunc

	PrintStack bool

	ExitAfterPrint bool
}

func (e *FinalError) Error() string {
	return e.Err.Error()
}

/*
@ctx:	插件服务端初始上下文（默认为pluginclient.Global_Context）

@err:	最终生成的error

@exit_after_print: 打印完错误链信息后是否结束主程序
*/
func ErrorTransmit(_severity string, _err error, _exit_after_print, _print_stack bool) {
	if ErrorManager == nil {
		logger.Error("globalerrormanager is nil")
		global.Close()
		os.Exit(1)
	}

	if _exit_after_print {
		ctx, cancel := context.WithCancel(ErrorManager.cancelCtx)
		ErrorManager.ErrCh <- &FinalError{
			Err:            _err,
			Cancel:         cancel,
			Severity:       _severity,
			PrintStack:     _print_stack,
			ExitAfterPrint: _exit_after_print,
		}
		<-ctx.Done()
		close(ErrorManager.ErrCh)
		global.Close()
		os.Exit(1)
	}

	ErrorManager.ErrCh <- &FinalError{
		Err:            _err,
		PrintStack:     _print_stack,
		ExitAfterPrint: _exit_after_print,
		Cancel:         nil,
	}
}
