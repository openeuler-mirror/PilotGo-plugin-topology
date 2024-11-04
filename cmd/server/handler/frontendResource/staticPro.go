//go:build production
// +build production

package frontendResource

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//go:embed assets index.html
var StaticFiles embed.FS

func StaticRouter(router *gin.Engine) {
	sf, err := fs.Sub(StaticFiles, "assets")
	if err != nil {
		err = errors.New(err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		return
	}

	mime.AddExtensionType(".js", "application/javascript")
	static := router.Group("/plugin/topology")
	{
		static.StaticFS("/assets", http.FS(sf))
		static.GET("/", func(c *gin.Context) {
			c.FileFromFS("/", http.FS(StaticFiles))
		})

	}

	// 解决页面刷新404的问题
	router.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.RequestURI, "/plugin/topology/api") {
			c.FileFromFS("/", http.FS(StaticFiles))
			return
		}
		c.AbortWithStatus(http.StatusNotFound)
	})

}
