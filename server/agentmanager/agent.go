package agentmanager

import (
	"sync"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
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

	Host_2             *meta.Host            `json:"host"`
	Processes_2        []*meta.Process       `json:"processes"`
	Netconnections_2   []*meta.Netconnection `json:"netconnections"`
	NetIOcounters_2    []*meta.NetIOcounter  `json:"netiocounters"`
	AddrInterfaceMap_2 map[string][]string   `json:"addrinterfacemap"`
	Disks_2            []*meta.Disk          `json:"disks"`
	Cpus_2             []*meta.Cpu           `json:"cpus"`
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
		err := errors.Errorf("delete unknown agent:%s **errstack**2", uuid) // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, false)
	}
}

func (am *AgentManager) AddAgent_T(a *Agent) {
	if a == nil {
		err := errors.Errorf("failed to add agent_t: %v **errstack**0", a) // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, false)
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
		err := errors.Errorf("delete unknown agent:%s **errstack**2", uuid) // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, false)
	}
}
