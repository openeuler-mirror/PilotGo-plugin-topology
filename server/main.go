package main

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/handler"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/logger"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/signal"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/service"
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
	*/
	db.InitDB()

	/*
		topo插件自身数据采集模块周期性数据采集: 全局网络拓扑、单机拓扑
	*/
	// ttcode: 测试自定义拓扑采集，临时注释
	service.PeriodCollectWorking([]string{}, [][]mysqlmanager.Filter_rule{})

	/*
		终止进程信号监听
	*/
	signal.SignalMonitoring()
}
