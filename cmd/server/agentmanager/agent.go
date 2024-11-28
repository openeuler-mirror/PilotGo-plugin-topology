/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package agentmanager

import (
	"sync"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/graph"
	"github.com/pkg/errors"
)

var Global_AgentManager *AgentManager

type AgentManager struct {
	PAgentMap sync.Map
	TAgentMap sync.Map

	AgentPort string
}

type Agent struct {
	ID         uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	UUID       string `gorm:"not null;unique" json:"uuid"`
	IP         string `gorm:"not null" json:"IP"`
	Port       string `gorm:"not null" json:"port"`
	Departid   string `json:"departid"`
	Departname string `json:"departname"`
	State      int    `gorm:"not null" json:"state"`
	TAState    int    `json:"TAstate"` // topo agent state: true(running) false(not runnings)

	Host_2             *graph.Host            `json:"host"`
	Processes_2        []*graph.Process       `json:"processes"`
	Netconnections_2   []*graph.Netconnection `json:"netconnections"`
	NetIOcounters_2    []*graph.NetIOcounter  `json:"netiocounters"`
	AddrInterfaceMap_2 map[string][]string    `json:"addrinterfacemap"`
	Disks_2            []*graph.Disk          `json:"disks"`
	Cpus_2             []*graph.Cpu           `json:"cpus"`
}

func InitAgentManager() {
	Global_AgentManager = &AgentManager{
		AgentPort: conf.Global_Config.Topo.Agent_port,
	}
}

func (am *AgentManager) AddAgent_P(a *Agent) {
	am.PAgentMap.Store(a.UUID, a)
}

func (am *AgentManager) GetAgent_P(uuid string) *Agent {
	if uuid != "" {
		if agent, ok := am.PAgentMap.Load(uuid); ok {
			return agent.(*Agent)
		}
	}

	return nil
}

func (am *AgentManager) DeleteAgent_P(uuid string) {
	if _, ok := am.PAgentMap.LoadAndDelete(uuid); !ok {
		err := errors.Errorf("delete unknown agent:%s", uuid)
		global.ERManager.ErrorTransmit("agentmanager", "error", err, false, true)
	}
}

func (am *AgentManager) AddAgent_T(a *Agent) {
	if a == nil {
		err := errors.Errorf("failed to add agent_t: %+v", a)
		global.ERManager.ErrorTransmit("agentmanager", "error", err, false, true)
		return
	}
	am.TAgentMap.Store(a.UUID, a)
}

func (am *AgentManager) GetAgent_T(uuid string) *Agent {
	if uuid == "" {
		return nil
	}

	agent, ok := am.TAgentMap.Load(uuid)
	if !ok {
		return nil
	}

	return agent.(*Agent)
}

func (am *AgentManager) DeleteAgent_T(uuid string) {
	if _, ok := am.TAgentMap.LoadAndDelete(uuid); !ok {
		err := errors.Errorf("delete unknown agent:%s", uuid)
		global.ERManager.ErrorTransmit("agentmanager", "error", err, false, true)
	}
}
