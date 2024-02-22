package handler

import (
	"fmt"
	"net/http"

	"gitee.com/openeuler/PilotGo-plugin-topology-agent/service"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Raw_metric_data(ctx *gin.Context) {
	data, err := service.DataCollectorService()
	if err != nil {
		err = errors.Wrap(err, "**2")
		fmt.Printf("%+v\n", err) // err top
		// errors.EORE(err)

		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": fmt.Errorf("(Raw_metric_data->DataCollectorService: %s)", err),
			"data":  nil,
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":  0,
		"error": nil,
		"data":  data,
	})
}
