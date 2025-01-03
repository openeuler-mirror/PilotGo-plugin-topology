/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Thu Nov 7 15:34:16 2024 +0800
 */
package webserver

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/timeout"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/webserver/frontendResource"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/webserver/handle"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/webserver/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func InitWebServer() {
	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil")
		global.ERManager.ErrorTransmit("webserver", "error", err, true, true)
		return
	}

	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.Logger([]string{
		"/plugin/topology/api/heartbeat",
		"/plugin_manage/bind",
		"/",
	}))
	gin.SetMode(gin.ReleaseMode)
	pluginclient.Global_Client.RegisterHandlers(engine)
	InitRouter(engine)
	frontendResource.StaticRouter(engine)

	web := &http.Server{
		Addr:    conf.Global_Config.Topo.Addr,
		Handler: engine,
	}

	global.ERManager.Wg.Add(1)
	go func() {
		if conf.Global_Config.Topo.Https_enabled {
			if err := web.ListenAndServeTLS(conf.Global_Config.Topo.Public_certificate, conf.Global_Config.Topo.Private_key); err != nil {
				if strings.Contains(err.Error(), "Server closed") {
					err = errors.New(err.Error())
					global.ERManager.ErrorTransmit("webserver", "info", err, false, false)
					return
				}
				err = errors.Errorf("%s, addr: %s", err.Error(), conf.Global_Config.Topo.Addr)
				global.ERManager.ErrorTransmit("webserver", "error", err, true, true)
			}
		}
		if err := web.ListenAndServe(); err != nil {
			if strings.Contains(err.Error(), "Server closed") {
				err = errors.New(err.Error())
				global.ERManager.ErrorTransmit("webserver", "info", err, false, false)
				return
			}
			err = errors.New(err.Error())
			global.ERManager.ErrorTransmit("webserver", "error", err, true, true)
		}
	}()

	go func() {
		defer global.ERManager.Wg.Done()

		<-global.ERManager.GoCancelCtx.Done()
		global.ERManager.ErrorTransmit("webserver", "info", errors.New("shutting down web server..."), false, false)

		ctx, cancel := context.WithTimeout(global.RootContext, 1*time.Second)
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
		api.POST("/heartbeat", handle.HeartbeatHandle)
		api.GET("/timestamps", handle.TimestampsHandle)
		api.GET("/agentlist", handle.AgentListHandle)
		api.GET("/batch_list", handle.BatchListHandle)
		api.GET("/batch_uuid", handle.BatchMachineListHandle)
	}

	collect := router.Group("/plugin/topology/api")
	{
		collect.POST("/deploy_collect_endpoint", handle.DeployCollectEndpointHandle)
		collect.POST("/collect_endpoint", handle.CollectEndpointHandle)
	}

	custom := router.Group("/plugin/topology/api")
	{
		custom.GET("/custom_topo_list", handle.CustomTopoListHandle)
		custom.POST("/create_custom_topo", handle.CreateCustomTopoHandle)
		custom.DELETE("/delete_custom_topo", handle.DeleteCustomTopoHandle)
		custom.PUT("/update_custom_topo", handle.UpdateCustomTopoHandle)
	}

	public := router.Group("/plugin/topology/api")
	{
		// public.GET("/single_host/:uuid", SingleHostHandle)
		public.GET("/single_host_tree/:uuid", handle.SingleHostTreeHandle)
		public.GET("/multi_host", handle.MultiHostHandle)
	}

	timeoutapi := router.Group("/plugin/topology/api")
	timeoutapi.Use(TimeoutMiddleware2(15 * time.Second))
	{
		timeoutapi.GET("/run_custom_topo", handle.RunCustomTopoHandle)
	}
}

func TimeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(12*time.Second),
		timeout.WithHandler(func(ctx *gin.Context) {
			ctx.Next()
		}),
		timeout.WithResponse(func(ctx *gin.Context) {
			ctx.JSON(http.StatusGatewayTimeout, gin.H{
				"code": http.StatusGatewayTimeout,
				"msg":  "server response timeout",
				"data": nil,
			})
		}),
	)
}

// 服务器响应超时中间件
func TimeoutMiddleware2(timeout time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {
		// 用超时context wrap request的context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer func() {
			// 检查是否超时
			if !c.GetBool("write") && ctx.Err() == context.DeadlineExceeded {
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			}
			//清理资源
			cancel()
		}()
		// 替换
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
