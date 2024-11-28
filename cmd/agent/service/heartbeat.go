/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/global"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/pkg/errors"
)

func SendHeartbeat() {
	agent_addr := conf.Config().Topo.Agent_addr

	global.ERManager.Wg.Add(1)
	go func() {
		global.ERManager.Wg.Done()
		for {
			select {
			case <-global.ERManager.GoCancelCtx.Done():
				return
			default:
				err := sendHeartbeat(agent_addr)
				if err != nil {
					err = errors.Wrap(err, " ")
					global.ERManager.ErrorTransmit("service", "error", err, false, false)
				}
				time.Sleep(time.Duration(conf.Config().Topo.Heartbeat) * time.Second)
			}
		}
	}()
}

func sendHeartbeat(addr string) error {
	m_u_bytes, err := global.FileReadBytes(global.Agentuuid_filepath)
	if err != nil {
		return errors.Wrap(err, " ")
	}
	type machineuuid struct {
		Agentuuid string `json:"agent_uuid"`
	}
	m_u_struct := &machineuuid{}
	json.Unmarshal(m_u_bytes, m_u_struct)

	type AgentHeartbeat struct {
		UUID              string
		Addr              string
		HeartbeatInterval int
		Time              time.Time
	}
	url := fmt.Sprintf("http://%s/plugin/topology/api/heartbeat", conf.Config().Topo.Server_addr)
	body := AgentHeartbeat{
		UUID:              m_u_struct.Agentuuid,
		Addr:              addr,
		HeartbeatInterval: conf.Config().Topo.Heartbeat,
	}
	params := &httputils.Params{
		Body: body,
	}

	resp, err := httputils.Post(url, params)
	if err != nil {
		return errors.Errorf("failed to send heartbeat: %s", err.Error())
	}

	// ttcode
	// url := fmt.Sprintf("http://%s/plugin/topology/api/heartbeat?uuid=%s&addr=%s&interval=%d", conf.Config().Topo.Server_addr, m_u_struct.Agentuuid, addr, conf.Config().Topo.Heartbeat)
	// resp, err := httputils.Get(url, nil)
	// if err != nil {
	// 	err = errors.Errorf("failed to send heartbeat: %s", err.Error())
	// 	return err
	// }

	resp_body := &struct {
		Code  int         `json:"code"`
		Error string      `json:"error"`
		Data  interface{} `json:"data"`
	}{}

	if len(resp.Body) == 0 {
		return errors.New("heartbeat resp.body is nil")
	}

	err = json.Unmarshal(resp.Body, resp_body)
	if err != nil {
		return errors.Errorf("failed to unmarshal json data: %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK || resp_body.Code != 0 {
		return errors.Errorf("failed to send heartbeat: url => %s, statuscode => %d, code => %d", url, resp.StatusCode, resp_body.Code)
	}

	return nil
}
