package handler

import (
	"github.com/gin-gonic/gin"
)

func InitRouter(router *gin.Engine) {
	api := router.Group("/plugin/api")
	{
		api.GET("/rawdata", Raw_metric_data)
	}
}
