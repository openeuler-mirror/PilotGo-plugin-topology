package main

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/handler"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/logger"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/service"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/signal"
	// "github.com/pyroscope-io/pyroscope/pkg/agent/profiler"
)

func main() {
	// profiler.Start(profiler.Config{
	// 	ApplicationName: "topo-server",
	// 	ServerAddress:   "http://localhost:4040",
	// })

	fmt.Println("hello topology")

	/*
		init config
		main.go中第一个执行的初始化模块，省略Global_Config全局指针变量调用的非空检测
	*/
	conf.InitConfig()

	/*
		init plugin client
	*/
	pluginclient.InitPluginClient()

	/*
		init error control
	*/
	errormanager.InitErrorManager()

	/*
		init agent manager
	*/
	agentmanager.InitAgentManager()

	/*
		init web server
	*/
	handler.InitWebServer()

	/*
		init logger
	*/
	logger.InitLogger()

	/*
		init machine agent list
	*/
	agentmanager.Global_AgentManager.InitMachineList()

	/*
		init database
		neo4j mysql redis prometheus
	*/
	db.InitDB()

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
