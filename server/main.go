package main

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/handler"
)

func main() {
	fmt.Println("hello topology")

	/*
		init config
	*/
	agentmanager.Topo.InitConfig()

	/*
		init plugin client
	*/
	agentmanager.Topo.InitPluginClient()

	/*
		init error control
	*/
	agentmanager.Topo.InitErrorControl(agentmanager.Topo.ErrCh, agentmanager.Topo.ErrGroup)

	/*
		init logger
	*/
	agentmanager.Topo.InitLogger()

	/*
		init JanusGraph
		TODO: 图数据库
	*/
	agentmanager.Topo.InitJanusGraph()

	/*
		init machine agent list
		TODO: 实时更新machine agent、topo agent的状态
	*/
	agentmanager.Topo.InitMachineList()

	/*
		init topo agent status
	*/

	/*
		init web server
	*/
	handler.InitWebServer()
}
