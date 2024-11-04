//go:build !production
// +build !production

package frontendResource

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func StaticRouter(router *gin.Engine) {
	static := router.Group("/plugin/topology")
	{
		static.Static("/assets", "web/dist/assets")
		static.StaticFile("/", "web/dist/index.html")

		// 解决页面刷新404的问题
		router.NoRoute(func(c *gin.Context) {
			if !strings.HasPrefix(c.Request.RequestURI, "/plugin/topology/api") {
				c.File("web/dist/index.html")
				return
			}
			c.AbortWithStatus(http.StatusNotFound)
		})
	}
}
