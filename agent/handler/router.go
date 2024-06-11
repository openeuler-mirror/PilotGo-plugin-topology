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

		if err := engine.Run(conf.Config().Topo.Agent_addr); err != nil {
			logger.Fatal("failed to run web server")
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
