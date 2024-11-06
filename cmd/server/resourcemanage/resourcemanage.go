package resourcemanage

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
)

// var ERManager *ErrorReleaseManagement

type ResourceReleaseFunction func()

type FinalError struct {
	Err error

	Severity string

	Cancel context.CancelFunc

	PrintStack bool

	ExitAfterPrint bool
}

func (e *FinalError) Error() string {
	return fmt.Sprintf("%+v", e.Err)
}

type ErrorReleaseManagement struct {
	ErrChan chan error

	errEndChan chan struct{}

	// cancelCtx: 控制 ERManager 本身资源释放的上下文，子上下文：errortransmit
	cancelCtx    context.Context
	cancelFunc   context.CancelFunc

	// GoCancelCtx: ERManager 用于控制项目指定goroutine优雅退出的上下文
	GoCancelCtx  context.Context
	GoCancelFunc context.CancelFunc

	Wg sync.WaitGroup

	// releaseFunc: 项目整体资源释放回调函数
	releaseFunc ResourceReleaseFunction
}

func CreateErrorReleaseManager(_ctx context.Context, _releaseFunc ResourceReleaseFunction) (*ErrorReleaseManagement, error) {
	if _ctx == nil || _releaseFunc == nil {
		return nil, fmt.Errorf("context or closeFunc is nil")
	}

	ErrorM := &ErrorReleaseManagement{
		ErrChan:     make(chan error, 20),
		errEndChan:  make(chan struct{}),
		releaseFunc: _releaseFunc,
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
			logger.Info("error management stopped")
			return
		case _error := <-erm.ErrChan:
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
	close(erm.ErrChan)

	time.Sleep(100 * time.Millisecond)
}

/*
@severity: debug info warn error

@err:	最终生成的error

@exit_after_print: 打印完异常日志后是否结束主程序

@print_stack: 是否打印异常日志错误链，打印错误链时默认severity为error
*/
func (erm *ErrorReleaseManagement) ErrorTransmit(_severity string, _err error, _exit_after_print, _print_stack bool) {
	if _exit_after_print {
		ctx, cancel := context.WithCancel(erm.cancelCtx)
		erm.ErrChan <- &FinalError{
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

	erm.ErrChan <- &FinalError{
		Err:            _err,
		Cancel:         nil,
		Severity:       _severity,
		PrintStack:     _print_stack,
		ExitAfterPrint: _exit_after_print,
	}
}
