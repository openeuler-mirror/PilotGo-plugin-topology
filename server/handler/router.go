package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/resourcemanage"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func InitWebServer() {
	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil **errstackfatal**2") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	engine := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	pluginclient.Global_Client.RegisterHandlers(engine)
	InitRouter(engine)
	StaticRouter(engine)

	webserver := &http.Server{
		Addr:    conf.Global_Config.Topo.Addr,
		Handler: engine,
	}

	resourcemanage.ERManager.Wg.Add(1)
	go func() {
		defer resourcemanage.ERManager.Wg.Done()

		if conf.Global_Config.Topo.Https_enabled {
			if err := webserver.ListenAndServeTLS(conf.Global_Config.Topo.Addr, conf.Global_Config.Topo.Public_certificate, conf.Global_Config.Topo.Private_key); err != nil {
				err = errors.Errorf("%s, addr: %s **errstackfatal**2", err.Error(), conf.Global_Config.Topo.Addr) // err top
				errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
			}
		}
		if err := webserver.ListenAndServe(); err != nil {
			err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		}
	}()

	go func() {
		<-resourcemanage.ERManager.GoCancelCtx.Done()

		logger.Info("shutting down web server...")

		ctx, cancel := context.WithTimeout(global.RootContext, 1*time.Second)
		defer cancel()

		if err := webserver.Shutdown(ctx); err != nil {
			logger.Error("web server shutdown error: %s", err.Error())
		} else {
			logger.Info("web server stopped gracefully")
		}
	}()

}

func InitRouter(router *gin.Engine) {
	api := router.Group("/plugin/topology/api")
	{
		api.POST("/heartbeat", HeartbeatHandle)
		api.GET("/timestamps", TimestampsHandle)
		api.GET("/agentlist", AgentListHandle)
		api.GET("/batch_list", BatchListHandle)
		api.GET("/batch_uuid", BatchMachineListHandle)
	}

	collect := router.Group("/plugin/topology/api")
	{
		collect.POST("/deploy_collect_endpoint", DeployCollectEndpointHandle)
		collect.POST("/collect_endpoint", CollectEndpointHandle)
	}

	custom := router.Group("/plugin/topology/api")
	{
		custom.GET("/custom_topo_list", CustomTopoListHandle)
		custom.POST("/create_custom_topo", CreateCustomTopoHandle)
		custom.DELETE("/delete_custom_topo", DeleteCustomTopoHandle)
		custom.PUT("/update_custom_topo", UpdateCustomTopoHandle)
	}

	public := router.Group("/plugin/topology/api")
	{
		// public.GET("/single_host/:uuid", SingleHostHandle)
		public.GET("/single_host_tree/:uuid", SingleHostTreeHandle)
		public.GET("/multi_host", MultiHostHandle)
	}

	timeoutapi := router.Group("/plugin/topology/api")
	timeoutapi.Use(TimeoutMiddleware2(15 * time.Second))
	{
		timeoutapi.GET("/run_custom_topo", RunCustomTopoHandle)
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
