/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package container

import (
	"encoding/json"
	"fmt"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/global"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

func ContainerList() ([]docker.APIContainers, error) {
	if DOCKERD_PORT == "" {
		bytes, err := global.FileReadBytes(DOCKER_CONFIG)
		if err != nil {
			return nil, errors.Wrap(err, " ")
		}
		var daemoncontent struct {
			Hosts []string `json:"hosts"`
		}
		if err := json.Unmarshal(bytes, &daemoncontent); err != nil {
			return nil, errors.New(err.Error())
		}

		for _, host := range daemoncontent.Hosts {
			if strings.HasPrefix(host, "tcp") {
				DOCKERD_PORT = strings.Split(host, ":")[2]
				break
			}
		}
	}

	if DOCKERD_PORT == "" {
		return nil, errors.New("no dockerd port found")
	}

	client, err := docker.NewClient(fmt.Sprintf("tcp://127.0.0.1:%s", DOCKERD_PORT))
	if err != nil {
		return nil, errors.New(err.Error())
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return nil, errors.New(err.Error())
	}

	return containers, nil
}
