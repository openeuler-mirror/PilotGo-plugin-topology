package webserver

import (
	"context"
	"net/http"
	"strings"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/webserver/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func InitWebServer() {
	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.Logger([]string{
		"/plugin/topology/api/health",
		"/",
	}))
	gin.SetMode(gin.ReleaseMode)
	InitRouter(engine)

	web := &http.Server{
		Addr:    conf.Config().Topo.Agent_addr,
		Handler: engine,
	}

	global.ERManager.Wg.Add(1)
	go func() {
		if conf.Config().Topo.Https_enabled {
			if err := web.ListenAndServeTLS(conf.Config().Topo.Public_certificate, conf.Config().Topo.Private_key); err != nil {
				if strings.Contains(err.Error(), "Server closed") {
					global.ERManager.ErrorTransmit("webserver", "info", errors.New(err.Error()), false, false)
					return
				}
				global.ERManager.ErrorTransmit("webserver", "error", errors.Errorf("failed to run web server: %+v", err.Error()), true, false)
			}
		}
		if err := web.ListenAndServe(); err != nil {
			if strings.Contains(err.Error(), "Server closed") {
				global.ERManager.ErrorTransmit("webserver", "info", errors.New(err.Error()), false, false)
				return
			}
			global.ERManager.ErrorTransmit("webserver", "error", errors.Errorf("failed to run web server: %+v", err.Error()), true, false)
		}
	}()

	go func() {
		defer global.ERManager.Wg.Done()

		<-global.ERManager.GoCancelCtx.Done()

		global.ERManager.ErrorTransmit("webserver", "info", errors.New("shutting down web server..."), false, false)

		ctx, cancel := context.WithTimeout(global.RootCtx, 1*time.Second)
		defer cancel()

		if err := web.Shutdown(ctx); err != nil {
			global.ERManager.ErrorTransmit("webserver", "error", errors.Errorf("web server shutdown error: %s", err.Error()), false, false)
		} else {
			global.ERManager.ErrorTransmit("webserver", "info", errors.New("web server stopped"), false, false)
		}
	}()
}

func InitRouter(router *gin.Engine) {
	api := router.Group("/plugin/topology/api")
	{
		api.GET("/health", HealthCheckHandle)
		api.GET("/rawdata", RawMetricDataHandle)
		api.GET("/container_list", ContainerListHandle)
	}
}
