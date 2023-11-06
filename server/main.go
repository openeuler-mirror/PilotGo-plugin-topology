package main

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/handler"
	service "gitee.com/openeuler/PilotGo-plugin-topology-server/service/background"
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
	agentmanager.Topo.InitErrorControl(agentmanager.Topo.ErrCh, agentmanager.Topo.Errmu, agentmanager.Topo.ErrCond)

	/*
		init logger
	*/
	agentmanager.Topo.InitLogger()

	/*
		init machine agent list
		TODO: 实时更新machine agent、topo agent的状态
	*/
	agentmanager.Topo.InitMachineList()

	/*
		init database
	*/
	service.InitDB()

	/*
		init web server
	*/
	handler.InitWebServer()

	/*
		终止进程信号监听
	*/
	agentmanager.Topo.SignalMonitoring(dao.Neo4j.Driver)
}
