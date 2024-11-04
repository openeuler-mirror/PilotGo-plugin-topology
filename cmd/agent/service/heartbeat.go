package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/utils"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/pkg/errors"
)

func SendHeartbeat() {
	agent_addr := conf.Config().Topo.Agent_addr

	go func() {
		for {
			err := sendHeartbeat(agent_addr)
			if err != nil {
				err = errors.Wrap(err, " ")
				logger.Error(err.Error())
			}
			time.Sleep(time.Duration(conf.Config().Topo.Heartbeat) * time.Second)
		}
	}()
}

func sendHeartbeat(addr string) error {
	m_u_bytes, err := utils.FileReadBytes(utils.Agentuuid_filepath)
	if err != nil {
		// err = errors.New(err.Error())
		logger.Error(err.Error())
		return err
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
		err = errors.Errorf("failed to send heartbeat: %s", err.Error())
		return err
	}

	resp_body := &struct {
		Code  int         `json:"code"`
		Error string      `json:"error"`
		Data  interface{} `json:"data"`
	}{}
	err = json.Unmarshal(resp.Body, resp_body)
	if err != nil {
		err = errors.Errorf("failed to unmarshal json data: %s", err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK || resp_body.Code != 0 {
		err = errors.Errorf("failed to send heartbeat: url => %s, statuscode => %d, code => %d", url, resp.StatusCode, resp_body.Code)
		return err
	}

	return nil
}
