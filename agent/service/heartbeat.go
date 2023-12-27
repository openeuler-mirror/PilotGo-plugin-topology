package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-agent/utils"
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
				err = errors.Wrap(err, " ") // err top
				fmt.Printf("%+v\n", err)
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

	url := fmt.Sprintf("http://%s/plugin/topology/api/heartbeat?agentaddr=%s&uuid=%s&interval=%d", conf.Config().Topo.Server_addr, addr, m_u_struct.Agentuuid, conf.Config().Topo.Heartbeat)
	resp, err := httputils.Post(url, nil)
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
