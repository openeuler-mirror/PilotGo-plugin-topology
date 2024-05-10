package agentmanager

import (
	"fmt"
	"net/http"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"github.com/pkg/errors"
)

func WaitingForHandshake() {
	i := 0
	loop := []string{`*.....`, `.*....`, `..*...`, `...*..`, `....*.`, `.....*`}
	for {
		if pluginclient.Global_Client != nil && pluginclient.Global_Client.Server() != "" {
			break
		}
		fmt.Printf("\r Waiting for handshake with pilotgo server%s", loop[i])
		if i < 5 {
			i++
		} else {
			i = 0
		}
		time.Sleep(1 * time.Second)
	}
}

func Wait4TopoServerReady() {
	for {
		url := "http://" + conf.Global_Config.Topo.Server_addr + "/plugin_manage/info"
		resp, err := http.Get(url)
		if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// 初始化PAgentMap中的agent
func (am *AgentManager) InitMachineList() {
	Wait4TopoServerReady()

	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil **errstackfatal**2") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	pluginclient.Global_Client.Wait4Bind()

	machine_list, err := pluginclient.Global_Client.MachineList()
	if err != nil {
		err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
	}

	for _, m := range machine_list {
		p := &Agent{}
		p.UUID = m.UUID
		p.Departname = m.Department
		p.IP = m.IP
		p.TAState = 0
		am.AddAgent_P(p)
	}
}

// 更新PAgentMap中的agent
func (am *AgentManager) UpdateMachineList() {
	machine_list, err := pluginclient.Global_Client.MachineList()
	if err != nil {
		err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
	}

	am.PAgentMap.Range(func(key, value interface{}) bool {
		am.DeleteAgent_P(key.(string))
		return true
	})

	for _, m := range machine_list {
		p := &Agent{}
		p.UUID = m.UUID
		p.Departname = m.Department
		p.IP = m.IP
		p.TAState = 0
		am.AddAgent_P(p)
	}
}
