package utils

import (
	"encoding/json"
	"strconv"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	docker "github.com/fsouza/go-dockerclient"

	"github.com/pkg/errors"
)

func ContainerList(agent *agentmanager.Agent) ([]docker.APIContainers, error) {
	resp, err := httputils.Get("http://"+agent.IP+":"+agent.Port+"/plugin/topology/api/container_list", nil)
	if err != nil {
		return nil, errors.Errorf("get container list from agent %s failed: %s **errstack**0", agent.IP, err.Error())
	}

	if resp == nil || resp.StatusCode != 200 {
		return nil, errors.Errorf("get container list from agent %s failed: %+v **errstack**0", agent.IP, resp)
	}

	resp_body := struct {
		Code int                    `json:"code"`
		Data []docker.APIContainers `json:"data"`
		Msg  string                 `json:"msg"`
	}{}

	err = json.Unmarshal(resp.Body, &resp_body)
	if err != nil {
		return nil, errors.Errorf("json unmarshal from agent %s failed: %s **errstack**2", agent.IP, err.Error())
	}

	return resp_body.Data, nil
}

func ProcessMatching(agent *agentmanager.Agent, exename, cmdline, component string) bool {
	component_lower := strings.ToLower(component)
	cmdline_lower := strings.ToLower(cmdline)
	cmdline_lower_arr := strings.Split(cmdline_lower, " ")

	switch exename {
	case "java":
		mainclass_match := false
		match_count := 0

		for i := 1; i < len(cmdline_lower_arr); i++ {
			if cmdline_lower_arr[i] == "-jar" && strings.Contains(cmdline_lower_arr[i+1], component_lower) {
				return true
			}

			if !strings.HasPrefix(cmdline_lower_arr[i], "-") && !strings.HasPrefix(cmdline_lower_arr[i], "/") {
				if strings.Contains(cmdline_lower_arr[i], component_lower) {
					mainclass_match = true
				}
			}

			if strings.Contains(cmdline_lower_arr[i], component_lower) {
				match_count = match_count + 1
			}

			if mainclass_match && match_count >= 1 {
				return true
			}
		}
	case "python", "python3", "python2":
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if strings.HasSuffix(cmdline_lower_arr[i], ".py") && strings.Split(strings.Split(cmdline_lower_arr[i], "/")[len(strings.Split(cmdline_lower_arr[i], "/"))-1], ".py")[0] == component_lower {
				return true
			}
		}
	case "ruby":
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if strings.HasSuffix(cmdline_lower_arr[i], ".rb") && strings.Split(strings.Split(cmdline_lower_arr[i], "/")[len(strings.Split(cmdline_lower_arr[i], "/"))-1], ".rb")[0] == component_lower {
				return true
			}
		}
	case "node":
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if (strings.HasSuffix(cmdline_lower_arr[i], ".js") || strings.HasSuffix(cmdline_lower_arr[i], ".ts")) && strings.Split(strings.Split(cmdline_lower_arr[i], "/")[len(strings.Split(cmdline_lower_arr[i], "/"))-1], ".")[0] == component_lower {
				return true
			}
		}
	case "perl":
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if strings.HasSuffix(cmdline_lower_arr[i], ".pl") && strings.Split(strings.Split(cmdline_lower_arr[i], "/")[len(strings.Split(cmdline_lower_arr[i], "/"))-1], ".pl")[0] == component_lower {
				return true
			}
		}
	// ①组件名与容器名匹配；②进程命令行中的-container-port与容器信息中的port匹配
	case "docker-proxy":
		cmdline_container_port := ""
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if cmdline_lower_arr[i] == "-container-port" {
				cmdline_container_port = cmdline_lower_arr[i+1]
				break
			}
		}

		containers, err := ContainerList(agent)
		if err != nil {
			err = errors.Wrap(err, " **errstack**0") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			break
		}

		for _, container := range containers {
			for _, name := range container.Names {
				name_lower := strings.ToLower(name)
				if strings.HasPrefix(name_lower, "/") {
					name_lower = strings.Replace(name_lower, "/", "", -1)
				}
				if name_lower != component_lower {
					continue
				}
				for _, port := range container.Ports {
					if strconv.Itoa(int(port.PrivatePort)) == cmdline_container_port {
						return true
					}
				}
			}
		}
	case "nginx":
		if strings.ToLower(exename) == component_lower && strings.Contains(cmdline_lower, "master process") {
			return true
		}
	default:
		if strings.ToLower(exename) == component_lower {
			return true
		}
	}

	return false
}
