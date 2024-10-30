package errormanager

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
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
			if topoerr.Err != nil {
				errarr := strings.Split(errors.Cause(topoerr.Err).Error(), "**")
				if len(errarr) < 2 {
					logger.Error("topoerror type required in root error (err: %+v)", topoerr.Err)
					os.Exit(1)
				}

				switch errarr[1] {
				// 只打印最底层error的message，不展开错误链的调用栈
				case "debug":
					logger.Debug("%+v\n", strings.Split(errors.Cause(topoerr.Err).Error(), "**")[0])
				// 只打印最底层error的message，不展开错误链的调用栈
				case "warn":
					logger.Warn("%+v\n", strings.Split(errors.Cause(topoerr.Err).Error(), "**")[0])
				// 打印错误链的调用栈
				case "errstack":
					fmt.Fprintf(Global_ErrorManager.Out, "%+v\n", topoerr.Err)
					// errors.EORE(err)
				// 打印错误链的调用栈，并结束程序
				case "errstackfatal": 
					fmt.Fprintf(Global_ErrorManager.Out, "%+v\n", topoerr.Err)
					// errors.EORE(err)
					topoerr.Cancel()
				default:
					fmt.Printf("only support \"debug warn errstack errstackfatal\" error type: %+v\n", topoerr.Err)
					os.Exit(1)
				}
			}
		}
	}(Global_ErrorManager.ErrCh)
}
