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
	agentmanager.InitPluginClient()

	/*
		init error control
	*/
	agentmanager.Topo.InitErrorControl(agentmanager.Topo.ErrCh)

	/*
		init web server
	*/
	handler.InitWebServer()

	/*
		init logger
	*/
	agentmanager.Topo.InitLogger()

	/*
		init machine agent list
	*/
	agentmanager.Topo.InitMachineList()

	/*
		init database
	*/
	service.InitDB()

	/*
		topo插件自身数据采集模块周期性数据采集: 全局网络拓扑、单机拓扑
	*/
	service.PeriodCollectWorking([]string{})

	/*
		终止进程信号监听
	*/
	agentmanager.Topo.SignalMonitoring(dao.Global_Neo4j.Driver, dao.Global_redis.Client)
}
