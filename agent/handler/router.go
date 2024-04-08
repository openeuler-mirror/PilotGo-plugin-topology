package handler

import (
	"github.com/gin-gonic/gin"
)

func InitRouter(router *gin.Engine) {
	api := router.Group("/plugin/topology/api")
	{
		api.GET("/health", HealthCheckHandle)
		api.GET("/rawdata", RawMetricDataHandle)
		api.GET("/container_list", ContainerListHandle)
	}
}
