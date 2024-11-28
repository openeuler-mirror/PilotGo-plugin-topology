/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package main

import (
	"fmt"
	"os"
	"runtime"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/resourcemanage"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/service"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/signal"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/webserver"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/global"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func main() {
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

	if err := logger.Init(conf.Config().Logopts); err != nil {
		fmt.Printf("logger init failed, please check the config file: %s", err.Error())
		os.Exit(1)
	}

	ermanager, err := resourcemanage.CreateErrorReleaseManager(global.RootCtx, global.Close)
	if err != nil {
		logger.Fatal(err.Error())
	}
	global.ERManager = ermanager

	webserver.InitWebServer()

	service.SendHeartbeat()

	signal.SignalMonitoring()
}