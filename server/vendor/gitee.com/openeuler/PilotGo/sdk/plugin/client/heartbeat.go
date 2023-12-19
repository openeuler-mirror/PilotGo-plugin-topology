package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
)

var HeartbeatInterval = 30 * time.Second

func (client *Client) SendHeartbeat() {
	clientID := client.PluginInfo.Url + "+" + client.PluginInfo.Name
	go func() {
		for {
			err := client.sendHeartbeat(clientID)
			if err != nil {
				logger.Error("Heartbeat failed:%v", err)
			}
			time.Sleep(HeartbeatInterval)
		}
	}()
}

func (client *Client) sendHeartbeat(clientID string) error {
	p := &struct {
		ClientID string `json:"clientID"`
	}{
		ClientID: clientID,
	}

	ServerUrl := "http://" + client.Server() + "/api/v1/pluginapi/heartbeat"
	resp, err := httputils.Post(ServerUrl, &httputils.Params{
		Body: p,
	})
	if err != nil {
		return err
	}
	res := &struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}{}
	if err := json.Unmarshal(resp.Body, res); err != nil {
		return err
	}
	if res.Code != http.StatusOK {
		return fmt.Errorf("heartbeat failed with status: %v", res.Code)
	}
	return nil
}
