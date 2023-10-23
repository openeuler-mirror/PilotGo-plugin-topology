package agentmanager

import (
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"github.com/pkg/errors"
)

type Agent_m struct {
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

func (t *Topoclient) AddAgent(a *Agent_m) {
	t.AgentMap.Store(a.UUID, a)
}

func (t *Topoclient) GetAgent(uuid string) *Agent_m {
	agent, ok := t.AgentMap.Load(uuid)
	if ok {
		return agent.(*Agent_m)
	}
	return nil
}

func (t *Topoclient) DeleteAgent(uuid string) {
	if _, ok := t.AgentMap.LoadAndDelete(uuid); !ok {
		err := errors.Errorf("delete unknown agent:%s **warn**2", uuid) // err top
		t.ErrCh <- err
	}
}

// 获取运行状态agent的数目
func (t *Topoclient) GetRunningAgentNumber() int {
	var agent_count int

	agent_count = 0

	Topo.AgentMap.Range(func(key, value interface{}) bool {
		agent := value.(*Agent_m)
		if agent.Host_2 != nil {
			agent_count++
		}

		return true
	})

	return agent_count
}
