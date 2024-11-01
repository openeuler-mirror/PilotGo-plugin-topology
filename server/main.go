package main

import (
	"gitee.com/openeuler/PilotGo-plugin-topology/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/handler"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/logger"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/resourcemanage"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/service"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/signal"
	// "github.com/pyroscope-io/pyroscope/pkg/agent/profiler"
)

func main() {
	// profiler.Start(profiler.Config{
	// 	ApplicationName: "topo-server",
	// 	ServerAddress:   "http://localhost:4040",
	// })

	/*
		init config
		main.go中第一个执行的初始化模块，省略Global_Config全局指针变量调用的非空检测
	*/
	conf.InitConfig()

	/*
		init logger
	*/
	logger.InitLogger()

	/*
		init error control、resource release、goroutine end
	*/
	resourcemanage.InitResourceManage()

	/*
		init plugin client
	*/
	pluginclient.InitPluginClient()

	/*
		init agent manager
	*/
	agentmanager.InitAgentManager()

	/*
		init database
		neo4j mysql redis prometheus
	*/
	db.InitDB()

	/*
		init web server
	*/
	handler.InitWebServer()

	/*
		init machine agent list
	*/
	agentmanager.Global_AgentManager.InitMachineList()

	/*
		topo插件自身数据采集模块周期性数据采集: 全局网络拓扑、单机拓扑
	*/
	service.InitPeriodCollectWorking([]string{}, [][]mysqlmanager.Filter_rule{})

	/*
		终止进程信号监听
		main.go中最后执行，不进行全局指针变量非空检测
	*/
	signal.SignalMonitoring()
}
