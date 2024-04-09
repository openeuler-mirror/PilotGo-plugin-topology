package pluginclient

import (
	"context"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo/sdk/common"
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
)

var GlobalClient *client.Client

var GlobalContext context.Context

const Version = "1.0.1"

var PluginInfo = &client.PluginInfo{
	Name:        "topology",
	Version:     Version,
	Description: "System application architecture detection.",
	Author:      "wangjunqi",
	Email:       "wangjunqi@kylinos.cn",
	Url:         "http://10.1.10.131:9991",
	PluginType:  "micro-app",
}

func InitPluginClient() {
	PluginInfo.Url = "http://" + conf.Config().Topo.Server_addr
	GlobalClient = client.DefaultClient(PluginInfo)

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
	GlobalClient.RegisterExtention(ex)

	GlobalContext = context.Background()
}
