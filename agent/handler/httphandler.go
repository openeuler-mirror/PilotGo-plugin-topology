package handler

import (
	"fmt"
	"net/http"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/agent/service"
	"gitee.com/openeuler/PilotGo-plugin-topology/agent/service/container"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func RawMetricDataHandle(ctx *gin.Context) {
	// 验证topo server请求来源
	if ctx.RemoteIP() != strings.Split(conf.Config().Topo.Server_addr, ":")[0] {
		err := errors.Errorf("unknow topo server: %s", ctx.RemoteIP())
		logger.ErrorStack("", err)
		// errors.EORE(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	data, err := service.DataCollectorService()
	if err != nil {
		err = errors.Wrap(err, "**2")
		logger.ErrorStack("", err)
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

func HealthCheckHandle(ctx *gin.Context) {
	// 验证topo server请求来源
	if ctx.RemoteIP() != strings.Split(conf.Config().Topo.Server_addr, ":")[0] {
		err := errors.Errorf("unknow topo server: %s", ctx.RemoteIP())
		logger.ErrorStack("", err)
		// errors.EORE(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	agentinfo := struct {
		Interval int `json:"interval"`
	}{
		Interval: conf.Config().Topo.Heartbeat,
	}

	response.Success(ctx, agentinfo, fmt.Sprintf("agent %s is running", conf.Config().Topo.Agent_addr))
}

func ContainerListHandle(ctx *gin.Context) {
	// 验证topo server请求来源
	if ctx.RemoteIP() != strings.Split(conf.Config().Topo.Server_addr, ":")[0] {
		err := errors.Errorf("unknow topo server: %s", ctx.RemoteIP())
		logger.ErrorStack("", err)
		// errors.EORE(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	containers, err := container.ContainerList()
	if err != nil {
		err = errors.Wrap(err, "")
		logger.ErrorStack("", err)
		response.Fail(ctx, nil, errors.Cause(err).Error())
	}

	response.Success(ctx, containers, "success")
}