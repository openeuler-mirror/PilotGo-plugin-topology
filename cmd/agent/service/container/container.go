package container

import (
	"encoding/json"
	"fmt"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/utils"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

func ContainerList() ([]docker.APIContainers, error) {
	if DOCKERD_PORT == "" {
		bytes, err := utils.FileReadBytes(DOCKER_CONFIG)
		if err != nil {
			return nil, errors.Wrap(err, " **errstack**0")
		}
		var daemoncontent struct {
			Hosts []string `json:"hosts"`
		}
		if err := json.Unmarshal(bytes, &daemoncontent); err != nil {
			return nil, errors.Errorf("%s **errstack**0", err.Error())
		}

		for _, host := range daemoncontent.Hosts {
			if strings.HasPrefix(host, "tcp") {
				DOCKERD_PORT = strings.Split(host, ":")[2]
				break
			}
		}
	}

	if DOCKERD_PORT == "" {
		return nil, errors.Errorf("no dockerd port found **errstack**0")
	}

	client, err := docker.NewClient(fmt.Sprintf("tcp://127.0.0.1:%s", DOCKERD_PORT))
	if err != nil {
		return nil, errors.Errorf("%s **errstack**0", err.Error())
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return nil, errors.Errorf("%s **errstack**0", err.Error())
	}

	return containers, nil
}
