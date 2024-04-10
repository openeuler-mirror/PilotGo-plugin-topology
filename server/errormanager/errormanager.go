package errormanager

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
)

var Global_ErrorManager *ErrorManager

type ErrorManager struct {
	ErrCh chan *Topoerror

	Out io.Writer
}

func InitErrorManager() {
	Global_ErrorManager = &ErrorManager{
		ErrCh: make(chan *Topoerror, 20),
	}

	switch conf.Global_Config.Logopts.Driver {
	case "stdout":
		Global_ErrorManager.Out = os.Stdout
	case "file":
		logfile, err := os.OpenFile(conf.Global_Config.Logopts.Path, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		Global_ErrorManager.Out = logfile
	}

	go func(ch <-chan *Topoerror) {
		for topoerr := range ch {
			// topoerr, ok := <-ch
			// if !ok {
			// 	break
			// }

			if topoerr.Err != nil {
				errarr := strings.Split(errors.Cause(topoerr.Err).Error(), "**")
				if len(errarr) < 2 {
					logger.Error("topoerror type required in root error (err: %+v)", topoerr.Err)
					os.Exit(1)
				}

				switch errarr[1] {
				case "debug": // 只打印最底层error的message，不展开错误链的调用栈
					logger.Debug("%+v\n", strings.Split(errors.Cause(topoerr.Err).Error(), "**")[0])
				case "warn": // 只打印最底层error的message，不展开错误链的调用栈
					logger.Warn("%+v\n", strings.Split(errors.Cause(topoerr.Err).Error(), "**")[0])
				case "errstack": // 打印错误链的调用栈
					fmt.Fprintf(Global_ErrorManager.Out, "%+v\n", topoerr.Err)
					// errors.EORE(err)
				case "errstackfatal": // 打印错误链的调用栈，并结束程序
					fmt.Fprintf(Global_ErrorManager.Out, "%+v\n", topoerr.Err)
					// errors.EORE(err)
					topoerr.Cancel()
				default:
					fmt.Printf("only support \"info warn fatal\" error type: %+v\n", topoerr.Err)
					os.Exit(1)
				}
			}
		}
	}(Global_ErrorManager.ErrCh)
}
