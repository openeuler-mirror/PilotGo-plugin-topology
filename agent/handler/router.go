package handler

import (
	"gitee.com/openeuler/PilotGo-plugin-topology/agent/conf"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/gin-gonic/gin"
)

func InitWebServer() {
	go func() {
		engine := gin.Default()
		gin.SetMode(gin.ReleaseMode)
		InitRouter(engine)

		if conf.Config().Topo.Https_enabled {
			err := engine.RunTLS(conf.Config().Topo.Agent_addr, conf.Config().Topo.Public_certificate, conf.Config().Topo.Private_key)
			if err != nil {
				logger.Fatal("failed to run web server: %+v", err.Error())
			}
		} else {
			err := engine.Run(conf.Config().Topo.Agent_addr)
			if err != nil {
				logger.Fatal("failed to run web server: %+v", err.Error())
			}
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
