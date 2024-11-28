/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package agentmanager

import (
	"fmt"
	"net/http"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
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
	defer global.ERManager.Wg.Done()
	global.ERManager.Wg.Add(1)
	for {
		select {
		case <-global.ERManager.GoCancelCtx.Done():
			return
		default:
			url := "http://" + conf.Global_Config.Topo.Addr + "/plugin_manage/info"
			resp, err := httputils.Get(url, nil)
			if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// 初始化PAgentMap中的agent
func (am *AgentManager) InitMachineList() {
	Wait4TopoServerReady()

	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil")
		global.ERManager.ErrorTransmit("agentmanager", "error", err, true, true)
		return
	}

	pluginclient.Global_Client.Wait4Bind()

	machine_list, err := pluginclient.Global_Client.MachineList()
	if err != nil {
		err = errors.Errorf(err.Error())
		global.ERManager.ErrorTransmit("agentmanager", "error", err, true, true)
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
		err = errors.Errorf(err.Error())
		global.ERManager.ErrorTransmit("agentmanager", "error", err, true, true)
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
