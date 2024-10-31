package resourcemanage

import (
	"context"
	"fmt"
	"os"
	"sync"

	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
)

var ERManager *ErrorReleaseManagement

type ResourceReleaseFunction func()

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

type ErrorReleaseManagement struct {
	errChan chan error

	cancelCtx    context.Context
	cancelFunc   context.CancelFunc
	GoCancelCtx  context.Context
	GoCancelFunc context.CancelFunc

	Wg sync.WaitGroup

	errEndChan chan struct{}

	releaseFunc ResourceReleaseFunction
}

func CreateErrorReleaseManager(_ctx context.Context, _releaseFunc ResourceReleaseFunction) (*ErrorReleaseManagement, error) {
	if _ctx == nil || _releaseFunc == nil {
		return nil, fmt.Errorf("context or closeFunc is nil")
	}

	ErrorM := &ErrorReleaseManagement{
		errChan:    make(chan error, 20),
		errEndChan: make(chan struct{}),
		releaseFunc:  _releaseFunc,
	}
	ErrorM.cancelCtx, ErrorM.cancelFunc = context.WithCancel(_ctx)
	ErrorM.GoCancelCtx, ErrorM.GoCancelFunc = context.WithCancel(_ctx)

	go ErrorM.errorFactory()

	return ErrorM, nil
}

func (erm *ErrorReleaseManagement) errorFactory() {
	for {
		select {
		case <-erm.errEndChan:
			logger.Info("errormanager exit")
			return
		case _error := <-erm.errChan:
			_terror, ok := _error.(*FinalError)
			if !ok {
				logger.Error("plain error: %s", _error.Error())
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
}

func (erm *ErrorReleaseManagement) ResourceRelease() {
	erm.releaseFunc()

	erm.GoCancelFunc()

	erm.Wg.Wait()

	close(erm.errEndChan)

	close(erm.errChan)
}

/*
@ctx:	插件服务端初始上下文（默认为pluginclient.Global_Context）

@err:	最终生成的error

@exit_after_print: 打印完错误链信息后是否结束主程序
*/
func (erm *ErrorReleaseManagement) ErrorTransmit(_severity string, _err error, _exit_after_print, _print_stack bool) {
	if _exit_after_print {
		ctx, cancel := context.WithCancel(erm.cancelCtx)
		erm.errChan <- &FinalError{
			Err:            _err,
			Cancel:         cancel,
			Severity:       _severity,
			PrintStack:     _print_stack,
			ExitAfterPrint: _exit_after_print,
		}
		<-ctx.Done()
		erm.ResourceRelease()
		os.Exit(1)
	}

	erm.errChan <- &FinalError{
		Err:            _err,
		PrintStack:     _print_stack,
		ExitAfterPrint: _exit_after_print,
		Cancel:         nil,
	}
}
