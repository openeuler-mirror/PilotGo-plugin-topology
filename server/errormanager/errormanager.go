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

var ErrorManager *ErrorManagement

type ErrorManagement struct {
	ErrCh chan error

	out io.Writer

	cancelCtx  context.Context
	cancelFunc context.CancelFunc
}

func CreateErrorManager() {
	ErrorManager = &ErrorManagement{
		ErrCh: make(chan error, 20),
	}

	ErrorManager.cancelCtx, ErrorManager.cancelFunc = context.WithCancel(global.Global_cancelCtx)

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

	global.Global_wg.Add(1)
	go func(ch <-chan error) {
		defer global.Global_wg.Done()
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
