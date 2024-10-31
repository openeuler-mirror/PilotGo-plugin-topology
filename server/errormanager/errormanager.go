package errormanager

import (
	"context"
	"fmt"
	"io"
	"os"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
)

type ResourceReleaseIface interface {
	Close()
}

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

var ErrorManager *ErrorManagement

type ErrorManagement struct {
	ErrCh chan error

	out io.Writer

	cancelCtx  context.Context
	cancelFunc context.CancelFunc

	end ResourceReleaseIface
}

func CreateErrorManager(_end ResourceReleaseIface) {
	ErrorManager = &ErrorManagement{
		ErrCh: make(chan error, 20),
		end:   _end,
	}
	ErrorManager.cancelCtx, ErrorManager.cancelFunc = context.WithCancel(global.END.CancelCtx)

	switch conf.Global_Config.Logopts.Driver {
	case "stdout":
		ErrorManager.out = os.Stdout
	case "file":
		logfile, err := os.OpenFile(conf.Global_Config.Logopts.Path, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		ErrorManager.out = logfile
	}

	global.END.Wg.Add(1)
	go func(ch <-chan error) {
		defer global.END.Wg.Done()
		for {
			select {
			case <-ErrorManager.cancelCtx.Done():
				return
			case _error := <-ch:
				_terror, ok := _error.(*FinalError)
				if !ok {
					fmt.Fprintf(ErrorManager.out, "%+v\n", _error)
					continue
				}

				if _terror.Err != nil {
					if !_terror.PrintStack && !_terror.ExitAfterPrint {
						switch _terror.Severity {
						case "debug":
							logger.Debug(errors.Cause(_terror.Err).Error())
						case "info":
							logger.Info(errors.Cause(_terror.Err).Error())
						case "warn":
							logger.Warn(errors.Cause(_terror.Err).Error())
						case "error":
							logger.Error(errors.Cause(_terror.Err).Error())
						default:
							logger.Error("only support \"debug info warn error\" type: %s\n", errors.Cause(_terror.Err).Error())
						}
					} else if _terror.PrintStack && !_terror.ExitAfterPrint {
						logger.ErrorStack("%+v", _terror.Err)
						// errors.EORE(err)
					} else if !_terror.PrintStack && _terror.ExitAfterPrint {
						switch _terror.Severity {
						case "debug":
							logger.Debug(errors.Cause(_terror.Err).Error())
						case "info":
							logger.Info(errors.Cause(_terror.Err).Error())
						case "warn":
							logger.Warn(errors.Cause(_terror.Err).Error())
						case "error":
							logger.Error(errors.Cause(_terror.Err).Error())
						default:
							logger.Error("only support \"debug info warn error\" type: %s\n", errors.Cause(_terror.Err).Error())
						}
						_terror.Cancel()
					} else if _terror.PrintStack && _terror.ExitAfterPrint {
						logger.ErrorStack("%+v", _terror.Err)
						// errors.EORE(err)
						_terror.Cancel()
					}
				}
			}
		}
	}(ErrorManager.ErrCh)
}

/*
@ctx:	插件服务端初始上下文（默认为pluginclient.Global_Context）

@err:	最终生成的error

@exit_after_print: 打印完错误链信息后是否结束主程序
*/
func ErrorTransmit(_severity string, _err error, _exit_after_print, _print_stack bool) {
	if ErrorManager == nil {
		logger.Error("globalerrormanager is nil")
		ErrorManager.end.Close()
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
		ErrorManager.end.Close()
		os.Exit(1)
	}

	ErrorManager.ErrCh <- &FinalError{
		Err:            _err,
		PrintStack:     _print_stack,
		ExitAfterPrint: _exit_after_print,
		Cancel:         nil,
	}
}
