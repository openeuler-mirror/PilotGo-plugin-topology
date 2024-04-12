package collector

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type DataCollector struct{}

func CreateDataCollector() *DataCollector {
	return &DataCollector{}
}

func (d *DataCollector) CollectInstantData() []error {
	start := time.Now()
	var wg sync.WaitGroup
	var errorlist []error
	var errorlist_rwlock sync.RWMutex

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil **errstackfatal**0") // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
		return nil
	}

	agentmanager.Global_AgentManager.TAgentMap.Range(
		func(key, value interface{}) bool {
			wg.Add(1)

			go func() {
				defer wg.Done()
				// ttcode
				temp_start := time.Now()
				agent := value.(*agentmanager.Agent)
				agent.Port = conf.Global_Config.Topo.Agent_port
				err := d.GetCollectDataFromTopoAgent(agent)
				if err != nil {
					errorlist_rwlock.Lock()
					errorlist = append(errorlist, errors.Wrapf(err, "%s**2", agent.IP))
					errorlist_rwlock.Unlock()
				}
				agentmanager.Global_AgentManager.AddAgent_T(agent)
				// ttcode
				temp_elapse := time.Since(temp_start)
				logger.Info("\033[32mtopo server 采集数据获取时间\033[0m: %s, %v, total\n", agent.UUID, temp_elapse)
			}()

			return true
		},
	)

	wg.Wait()

	elapse := time.Since(start)
	// fmt.Fprintf(agentmanager.Topo.Out, "\033[32mtopo server 采集数据获取时间\033[0m: %v\n", elapse)
	logger.Info("\033[32mtopo server 采集数据获取时间\033[0m: %v\n", elapse)

	if len(errorlist) != 0 {
		return errorlist
	}
	return nil
}

func (d *DataCollector) GetCollectDataFromTopoAgent(agent *agentmanager.Agent) error {
	url := "http://" + agent.IP + ":" + agent.Port + "/plugin/topology/api/rawdata"

	resp, err := httputils.Get(url, nil)
	if err != nil {
		return errors.Errorf("%s, %s **errstack**2", url, err.Error())
	}

	// ttcode
	tmpfile, _ := os.CreateTemp("", "response")
	defer os.Remove(tmpfile.Name())
	reader := bytes.NewReader(resp.Body)
	io.Copy(tmpfile, reader)
	fileInfo, _ := tmpfile.Stat()
	logger.Info("\033[32mtopo server 采集数据大小\033[0m: %s, %d kb\n", agent.UUID, fileInfo.Size()/1024)

	if statuscode := resp.StatusCode; statuscode != 200 {
		return errors.Errorf("%v, %s **errstack**2", resp.StatusCode, url)
	}

	results := struct {
		Code  int         `json:"code"`
		Error string      `json:"error"`
		Data  interface{} `json:"data"`
	}{}

	err = json.Unmarshal(resp.Body, &results)
	if err != nil {
		return errors.Errorf("%s **errstack**2", err.Error())
	}

	collectdata := &struct {
		Host_1             *meta.Host
		Processes_1        []*meta.Process
		Netconnections_1   []*meta.Netconnection
		NetIOcounters_1    []*meta.NetIOcounter
		AddrInterfaceMap_1 map[string][]string
		Disks_1            []*meta.Disk
		Cpus_1             []*meta.Cpu
	}{}
	mapstructure.Decode(results.Data, collectdata)

	agent.Host_2 = collectdata.Host_1
	agent.Processes_2 = collectdata.Processes_1
	agent.Netconnections_2 = collectdata.Netconnections_1
	agent.NetIOcounters_2 = collectdata.NetIOcounters_1
	agent.AddrInterfaceMap_2 = collectdata.AddrInterfaceMap_1
	agent.Disks_2 = collectdata.Disks_1
	agent.Cpus_2 = collectdata.Cpus_1

	return nil
}
