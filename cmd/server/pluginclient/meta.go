/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2.
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package pluginclient

import (
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
)

const Version = "1.0.1"

var PluginInfo = &client.PluginInfo{
	Name:        "topology",
	Version:     Version,
	Description: "System application architecture detection.",
	Author:      "wangjunqi",
	Email:       "wangjunqi@kylinos.cn",
	Url:         "", // 远端访问插件服务端的地址
	Icon:        "Help",
	MenuName:    "集群拓扑",
	PluginType:  "micro-app",
}
