package main

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/handler"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	service "gitee.com/openeuler/PilotGo-plugin-topology-server/service/background"
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
	agentmanager.Topo.InitConfig()

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
	agentmanager.Topo.InitLogger()

	/*
		init machine agent list
	*/
	agentmanager.GlobalAgentManager.InitMachineList()

	/*
		init database
	*/
	service.InitDB()

	/*
		topo插件自身数据采集模块周期性数据采集: 全局网络拓扑、单机拓扑
	*/
	// ttcode: 测试自定义拓扑采集，临时注释
	service.PeriodCollectWorking([]string{}, [][]meta.Filter_rule{})

	/*
		终止进程信号监听
	*/
	agentmanager.Topo.SignalMonitoring(dao.Global_Neo4j.Driver, dao.Global_redis.Client)
}
