package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/graph"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type TopoGenerator struct {
	Factory TopoInterface
}

func CreateTopoGenerator(trules []mysqlmanager.Tag_rule, nrules [][]mysqlmanager.Filter_rule) *TopoGenerator {
	_topogenerator := &TopoGenerator{}

	if len(nrules) != 0 {
		_topogenerator.Factory = &CustomTopo{
			Tagrules:         trules,
			Noderules:        nrules,
			Agent_node_count: new(int32),
		}
		return _topogenerator
	}

	_topogenerator.Factory = &PublicTopo{
		Agent_node_count: new(int32),
	}
	return _topogenerator
}

func (t *TopoGenerator) ProcessingData(agentnum int) (*graph.Nodes, *graph.Edges, []error, []error) {
	nodes := &graph.Nodes{
		Lock:         sync.Mutex{},
		Lookup:       make(map[string]*graph.Node, 0),
		LookupByType: make(map[string][]*graph.Node, 0),
		LookupByUUID: make(map[string][]*graph.Node, 0),
		Nodes:        make([]*graph.Node, 0),
	}
	edges := &graph.Edges{
		Lock:           sync.Mutex{},
		Lookup:         sync.Map{},
		Node_Edges_map: sync.Map{},
		Edges:          make([]*graph.Edge, 0),
	}

	var wg sync.WaitGroup
	var collect_errorlist []error
	var process_errorlist []error
	var process_errorlist_rwlock sync.RWMutex

	collect_errorlist = t.collectInstantData()
	if len(collect_errorlist) != 0 {
		for i, err := range collect_errorlist {
			collect_errorlist[i] = errors.Wrap(err, " ")
		}
	}

	start := time.Now()

	ctx1, cancel1 := context.WithCancel(global.ERManager.GoCancelCtx)
	go func(cancelfunc context.CancelFunc) {
		for {
			if atomic.LoadInt32(t.Factory.Return_Agent_node_count()) == int32(agentnum) {
				cancelfunc()
				break
			}
		}
	}(cancel1)

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil")
		global.ERManager.ErrorTransmit("generator", "error", err, true, true)
		return nil, nil, nil, nil
	}

	agentmanager.Global_AgentManager.TAgentMap.Range(
		func(key, value interface{}) bool {
			wg.Add(1)

			agent := value.(*agentmanager.Agent)

			go func(ctx context.Context, _agent *agentmanager.Agent, _nodes *graph.Nodes, _edges *graph.Edges) {
				defer wg.Done()

				if _agent.Host_2 != nil && _agent.Processes_2 != nil && _agent.Netconnections_2 != nil {
					err := t.Factory.CreateNodeEntities(_agent, _nodes)
					if err != nil {
						process_errorlist_rwlock.Lock()
						process_errorlist = append(process_errorlist, errors.Wrap(err, " "))
						process_errorlist_rwlock.Unlock()
					}

					<-ctx.Done()

					err = t.Factory.CreateEdgeEntities(_agent, _edges, _nodes)
					if err != nil {
						process_errorlist_rwlock.Lock()
						process_errorlist = append(process_errorlist, errors.Wrap(err, " "))
						process_errorlist_rwlock.Unlock()
					}

				}
			}(ctx1, agent, nodes, edges)

			return true
		},
	)
	wg.Wait()

	atomic.StoreInt32(t.Factory.Return_Agent_node_count(), int32(0))

	elapse := time.Since(start)
	global.ERManager.ErrorTransmit("generator", "info", errors.Errorf("采集数据处理时间: %v\n", elapse), false, false)

	return nodes, edges, collect_errorlist, process_errorlist
}

func (t *TopoGenerator) collectInstantData() []error {
	start := time.Now()
	var wg sync.WaitGroup
	var errorlist []error
	var errorlist_rwlock sync.RWMutex

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil")
		global.ERManager.ErrorTransmit("generator", "error", err, true, true)
		return nil
	}

	agentmanager.Global_AgentManager.TAgentMap.Range(
		func(key, value interface{}) bool {
			wg.Add(1)

			go func() {
				defer wg.Done()

				temp_start := time.Now()
				agent := value.(*agentmanager.Agent)
				agent.Port = conf.Global_Config.Topo.Agent_port
				err := t.getCollectDataFromTopoAgent(agent)
				if err != nil {
					errorlist_rwlock.Lock()
					errorlist = append(errorlist, errors.Wrap(err, agent.IP))
					errorlist_rwlock.Unlock()
				}
				agentmanager.Global_AgentManager.AddAgent_T(agent)

				temp_elapse := time.Since(temp_start)
				global.ERManager.ErrorTransmit("generator", "info", errors.Errorf("采集数据获取时间: %s, %v, total\n", agent.UUID, temp_elapse), false, false)
			}()

			return true
		},
	)

	wg.Wait()

	elapse := time.Since(start)
	global.ERManager.ErrorTransmit("generator", "info", errors.Errorf("采集数据获取时间: %v\n", elapse), false, false)

	if len(errorlist) != 0 {
		return errorlist
	}
	return nil
}

func (t *TopoGenerator) getCollectDataFromTopoAgent(agent *agentmanager.Agent) error {
	url := "http://" + agent.IP + ":" + agent.Port + "/plugin/topology/api/rawdata"

	resp, err := httputils.Get(url, nil)
	if err != nil {
		return errors.Errorf("%s, %s", url, err.Error())
	}

	// ttcode
	tmpfile, _ := os.CreateTemp("", "response")
	defer os.Remove(tmpfile.Name())
	reader := bytes.NewReader(resp.Body)
	io.Copy(tmpfile, reader)
	fileInfo, _ := tmpfile.Stat()
	global.ERManager.ErrorTransmit("generator", "info", errors.Errorf("采集数据大小: %s, %d kb\n", agent.UUID, fileInfo.Size()/1024), false, false)

	if statuscode := resp.StatusCode; statuscode != 200 {
		return errors.Errorf("%v, %s", resp.StatusCode, url)
	}

	results := struct {
		Code  int         `json:"code"`
		Error string      `json:"error"`
		Data  interface{} `json:"data"`
	}{}

	err = json.Unmarshal(resp.Body, &results)
	if err != nil {
		return errors.New(err.Error())
	}

	collectdata := &struct {
		Host_1             *graph.Host
		Processes_1        []*graph.Process
		Netconnections_1   []*graph.Netconnection
		NetIOcounters_1    []*graph.NetIOcounter
		AddrInterfaceMap_1 map[string][]string
		Disks_1            []*graph.Disk
		Cpus_1             []*graph.Cpu
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
