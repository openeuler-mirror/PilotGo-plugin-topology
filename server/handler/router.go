package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func InitWebServer() {
	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil **errstackfatal**2") // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
		return
	}

	go func() {
		engine := gin.Default()
		gin.SetMode(gin.ReleaseMode)
		pluginclient.Global_Client.RegisterHandlers(engine)
		InitRouter(engine)
		StaticRouter(engine)

		err := engine.Run(conf.Global_Config.Topo.Server_addr)
		if err != nil {
			err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
			errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
		}
	}()
}

func InitRouter(router *gin.Engine) {
	api := router.Group("/plugin/topology/api")
	{
		api.POST("/heartbeat", HeartbeatHandle)
		api.GET("/timestamps", TimestampsHandle)
		api.GET("/agentlist", AgentListHandle)
		api.GET("/custom_topo_list", CustomTopoListHandle)
		api.GET("/batch_list", BatchListHandle)
		api.GET("/batch_uuid", BatchMachineListHandle)

		// api.GET("/single_host/:uuid", SingleHostHandle)
		api.GET("/single_host_tree/:uuid", SingleHostTreeHandle)
		api.GET("/multi_host", MultiHostHandle)
		api.POST("/create_custom_topo", CreateCustomTopoHandle)
		api.DELETE("/delete_custom_topo", DeleteCustomTopoHandle)
		api.PUT("/update_custom_topo", UpdateCustomTopoHandle)

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
