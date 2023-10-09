package agentmanager

import (
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
)

const Version = "1.0.1"

var PluginInfo = &client.PluginInfo{
	Name:        "topology",
	Version:     Version,
	Description: "system application architecture perception",
	Author:      "wangjunqi",
	Email:       "wangjunqi@kylinos.cn",
	Url:         "http://10.1.10.131:9991/plugin/topology",
}
