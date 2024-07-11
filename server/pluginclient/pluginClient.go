package pluginclient

import (
	"context"
	"os"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo/sdk/common"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
)

var Global_Client *client.Client

var Global_Context context.Context

func InitPluginClient() {
	Global_Client = client.DefaultClient(PluginInfo)

	// 注册插件扩展点
	var ex []common.Extention
	pe1 := &common.PageExtention{
		Type:       common.ExtentionPage,
		Name:       "配置列表",
		URL:        "/topoList",
		Permission: "plugin.topology.page/menu",
	}
	pe2 := &common.PageExtention{
		Type:       common.ExtentionPage,
		Name:       "创建配置",
		URL:        "/customTopo",
		Permission: "plugin.topology.page/menu",
	}
	ex = append(ex, pe1, pe2)
	Global_Client.RegisterExtention(ex)

	Global_Context = context.Background()

	go uploadResource()
}

func uploadResource() {
	for !Global_Client.IsBind() {
		time.Sleep(100 * time.Millisecond)
	}
	
	dirPath := conf.Global_Config.Topo.Path
	filename_list := []string{}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Error("fail to read files: %s", err.Error())
		return
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename_list = append(filename_list, file.Name())
	}

	for _, filename := range filename_list {
		err := Global_Client.FileUpload(dirPath, filename)
		if err != nil {
			logger.Error("fail to upload file: %s", err.Error())
		}
	}
}
