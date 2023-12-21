package service

import (
	"fmt"
	"net/http"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-agent/conf"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/pkg/errors"
)

func SendHeartbeat() {
	agentID := conf.Config().Topo.Agent_addr

	go func() {
		for {
			err := sendHeartbeat(agentID)
			if err != nil {
				err = errors.Wrap(err, " ") // err top
				fmt.Printf("%+v\n", err)
			}
			time.Sleep(time.Duration(conf.Config().Topo.Heartbeat) * time.Second)
		}
	}()
}

func sendHeartbeat(agentid string) error {
	url := "http://" + conf.Config().Topo.Server_addr + "/plugin/topology/api/heartbeat?agentid=" + agentid
	resp, err := httputils.Post(url, nil)
	if err != nil {
		err = errors.Errorf("failed to send heartbeat: %s", err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err = errors.Errorf("failed to send heartbeat: url => %s, statuscode => %d", url, resp.StatusCode)
		return err
	}

	return nil
}
