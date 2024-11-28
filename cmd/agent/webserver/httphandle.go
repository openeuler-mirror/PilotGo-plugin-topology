/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Fri Nov 8 09:13:05 2024 +0800
 */
package webserver

import (
	"fmt"
	"net/http"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/service"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/service/container"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func RawMetricDataHandle(ctx *gin.Context) {
	// 验证topo server请求来源
	if ctx.RemoteIP() != strings.Split(conf.Config().Topo.Server_addr, ":")[0] {
		err := errors.Errorf("unknow client request from %s: %s", ctx.RemoteIP(), ctx.Request.URL)
		global.ERManager.ErrorTransmit("webserver", "error", err, false, false)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	data, err := service.DataCollectorService()
	if err != nil {
		err = errors.Wrap(err, " ")
		global.ERManager.ErrorTransmit("webserver", "error", err, false, true)
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
		err := errors.Errorf("unknow client request from %s: %s", ctx.RemoteIP(), ctx.Request.URL)
		global.ERManager.ErrorTransmit("webserver", "error", err, false, false)
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
		err := errors.Errorf("unknow client request from %s: %s", ctx.RemoteIP(), ctx.Request.URL)
		global.ERManager.ErrorTransmit("webserver", "error", err, false, false)
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
		global.ERManager.ErrorTransmit("webserver", "error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
	}

	response.Success(ctx, containers, "success")
}
