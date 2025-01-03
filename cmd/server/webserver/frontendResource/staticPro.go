/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Thu Nov 7 15:34:16 2024 +0800
 */
//go:build production
// +build production

package frontendResource

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//go:embed assets index.html
var StaticFiles embed.FS

func StaticRouter(router *gin.Engine) {
	sf, err := fs.Sub(StaticFiles, "assets")
	if err != nil {
		err = errors.New(err.Error())
		global.ERManager.ErrorTransmit("webserver", "error", err, false, true)
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
