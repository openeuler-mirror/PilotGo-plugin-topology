package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/collector"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/graph"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
)

type DataProcesser struct {
	agent_node_count int32
}

func CreateDataProcesser() *DataProcesser {
	return &DataProcesser{}
}

func (d *DataProcesser) ProcessData(agentnum int, tagrules []mysqlmanager.Tag_rule, noderules [][]mysqlmanager.Filter_rule) (*graph.Nodes, *graph.Edges, []error, []error) {
	nodes := &graph.Nodes{
		Lock:         sync.Mutex{},
		Lookup:       make(map[string]*graph.Node, 0),
		LookupByType: make(map[string][]*graph.Node, 0),
		LookupByUUID: make(map[string][]*graph.Node, 0),
		Nodes:        make([]*graph.Node, 0),
	}
	edges := &graph.Edges{
		Lock:      sync.Mutex{},
		Lookup:    sync.Map{},
		SrcToDsts: make(map[string][]string, 0),
		DstToSrcs: make(map[string][]string, 0),
		Edges:     make([]*graph.Edge, 0),
	}

	var wg sync.WaitGroup
	var collect_errorlist []error
	var process_errorlist []error
	var process_errorlist_rwlock sync.RWMutex

	datacollector := collector.CreateDataCollector()
	collect_errorlist = datacollector.CollectInstantData()
	if len(collect_errorlist) != 0 {
		for i, err := range collect_errorlist {
			collect_errorlist[i] = errors.Wrap(err, "**7")
		}
	}

	start := time.Now()

	ctx1, cancel1 := context.WithCancel(pluginclient.GlobalContext)
	go func(cancelfunc context.CancelFunc) {
		for {
			if atomic.LoadInt32(&d.agent_node_count) == int32(agentnum) {
				cancelfunc()
				break
			}
		}
	}(cancel1)

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil **errstackfatal**0") // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
		return nil, nil, nil, nil
	}

	agentmanager.Global_AgentManager.TAgentMap.Range(
		func(key, value interface{}) bool {
			wg.Add(1)

			agent := value.(*agentmanager.Agent)

			go func(ctx context.Context, _agent *agentmanager.Agent, _nodes *graph.Nodes, _edges *graph.Edges, _tagrules []mysqlmanager.Tag_rule, _noderules [][]mysqlmanager.Filter_rule) {
				defer wg.Done()

				if _agent.Host_2 != nil && _agent.Processes_2 != nil && _agent.Netconnections_2 != nil {
					if len(_noderules) != 0 {
						err := d.CustomCreateNodeEntities(_agent, _nodes, _tagrules, _noderules)
						if err != nil {
							process_errorlist_rwlock.Lock()
							process_errorlist = append(process_errorlist, errors.Wrap(err, "**2"))
							process_errorlist_rwlock.Unlock()
						}

						<-ctx.Done()

						err = d.CustomCreateEdgeEntities(_agent, _edges, _nodes)
						if err != nil {
							process_errorlist_rwlock.Lock()
							process_errorlist = append(process_errorlist, errors.Wrap(err, "**2"))
							process_errorlist_rwlock.Unlock()
						}
					} else {
						err := d.CreateNodeEntities(_agent, _nodes)
						if err != nil {
							process_errorlist_rwlock.Lock()
							process_errorlist = append(process_errorlist, errors.Wrap(err, "**2"))
							process_errorlist_rwlock.Unlock()
						}

						<-ctx.Done()

						err = d.CreateEdgeEntities(_agent, _edges, _nodes)
						if err != nil {
							process_errorlist_rwlock.Lock()
							process_errorlist = append(process_errorlist, errors.Wrap(err, "**2"))
							process_errorlist_rwlock.Unlock()
						}
					}
				}
			}(ctx1, agent, nodes, edges, tagrules, noderules)

			return true
		},
	)
	wg.Wait()

	atomic.StoreInt32(&d.agent_node_count, int32(0))

	elapse := time.Since(start)
	logger.Info("\033[32mtopo server 采集数据处理时间\033[0m: %v\n", elapse)

	return nodes, edges, collect_errorlist, process_errorlist
}

func (d *DataProcesser) CreateNodeEntities(agent *agentmanager.Agent, nodes *graph.Nodes) error {
	host_node := &graph.Node{
		ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_HOST, global.NODE_CONNECTOR, agent.IP),
		Name:       agent.UUID,
		Type:       global.NODE_HOST,
		UUID:       agent.UUID,
		LayoutAttr: global.INNER_LAYOUT_1,
		ComboId:    agent.UUID,
		Metrics:    *utils.HostToMap(agent.Host_2, &agent.AddrInterfaceMap_2),
	}

	nodes.Add(host_node)

	for _, process := range agent.Processes_2 {
		proc_node := &graph.Node{
			ID:         fmt.Sprintf("%s%s%s%s%d", agent.UUID, global.NODE_CONNECTOR, global.NODE_PROCESS, global.NODE_CONNECTOR, process.Pid),
			Name:       process.ExeName,
			Type:       global.NODE_PROCESS,
			UUID:       agent.UUID,
			LayoutAttr: global.INNER_LAYOUT_2,
			ComboId:    agent.UUID,
			Metrics:    *utils.ProcessToMap(process),
		}

		nodes.Add(proc_node)

		for _, thread := range process.Threads {
			thre_node := &graph.Node{
				ID:         fmt.Sprintf("%s%s%s%s%d", agent.UUID, global.NODE_CONNECTOR, global.NODE_THREAD, global.NODE_CONNECTOR, thread.Tid),
				Name:       strconv.Itoa(int(thread.Tid)),
				Type:       global.NODE_THREAD,
				UUID:       agent.UUID,
				LayoutAttr: global.INNER_LAYOUT_3,
				ComboId:    agent.UUID,
				Metrics:    *utils.ThreadToMap(&thread),
			}

			nodes.Add(thre_node)
		}

		// for _, net := range process.NetIOCounters {
		// 	net_node := &graph.Node{
		// 		ID:      fmt.Sprintf("%s-%s-%d", agent.UUID, global.NODE_NET, process.Pid),
		// 		Name:    net.Name,
		// 		Type:    global.NODE_NET,
		// 		UUID:    agent.UUID,
		// 		Metrics: *utils.NetToMap(&net, &agent.AddrInterfaceMap_2),
		// 	}

		// 	nodes.Add(net_node)
		// }
	}

	// 临时定义不含网络流量metric的网络节点
	for _, net := range agent.Netconnections_2 {
		if laddr_slice := strings.Split(net.Laddr, ":"); len(laddr_slice) != 0 {
			net_node := &graph.Node{
				ID:         fmt.Sprintf("%s%s%s%s%d:%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_NET, global.NODE_CONNECTOR, net.Pid, laddr_slice[1]),
				Name:       net.Laddr,
				Type:       global.NODE_NET,
				UUID:       agent.UUID,
				LayoutAttr: global.INNER_LAYOUT_5,
				ComboId:    agent.UUID,
				Metrics:    *utils.NetToMap(net),
			}

			nodes.Add(net_node)
		} else {
			err := errors.Errorf("syntax error: %s **errstack**13", net.Laddr) // err top
			errormanager.ErrorTransmit(pluginclient.GlobalContext, err, false)
		}
	}

	for _, disk := range agent.Disks_2 {
		disk_node := &graph.Node{
			ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_RESOURCE, global.NODE_CONNECTOR, disk.Partition.Device),
			Name:       disk.Partition.Device,
			Type:       global.NODE_RESOURCE,
			UUID:       agent.UUID,
			LayoutAttr: global.INNER_LAYOUT_4,
			ComboId:    agent.UUID,
			Metrics:    *utils.DiskToMap(disk),
		}

		nodes.Add(disk_node)
	}

	for _, cpu := range agent.Cpus_2 {
		cpu_node := &graph.Node{
			ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_RESOURCE, global.NODE_CONNECTOR, "CPU"+strconv.Itoa(int(cpu.Info.CPU))),
			Name:       "CPU" + strconv.Itoa(int(cpu.Info.CPU)),
			Type:       global.NODE_RESOURCE,
			UUID:       agent.UUID,
			LayoutAttr: global.INNER_LAYOUT_4,
			ComboId:    agent.UUID,
			Metrics:    *utils.CpuToMap(cpu),
		}

		nodes.Add(cpu_node)
	}

	for _, ifaceio := range agent.NetIOcounters_2 {
		iface_node := &graph.Node{
			ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_RESOURCE, global.NODE_CONNECTOR, "NC"+ifaceio.Name),
			Name:       "NC" + ifaceio.Name,
			Type:       global.NODE_RESOURCE,
			UUID:       agent.UUID,
			LayoutAttr: global.INNER_LAYOUT_4,
			ComboId:    agent.UUID,
			Metrics:    *utils.InterfaceToMap(ifaceio),
		}

		nodes.Add(iface_node)
	}

	atomic.AddInt32(&d.agent_node_count, int32(1))

	return nil
}

func (d *DataProcesser) CustomCreateNodeEntities(agent *agentmanager.Agent, nodes *graph.Nodes, tagrules []mysqlmanager.Tag_rule, noderules [][]mysqlmanager.Filter_rule) error {
	allconnections := []graph.Netconnection{}
	for _, net := range agent.Netconnections_2 {
		allconnections = append(allconnections, *net)
	}

	host_node := &graph.Node{
		ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_HOST, global.NODE_CONNECTOR, agent.IP),
		Name:       agent.UUID,
		Type:       global.NODE_HOST,
		UUID:       agent.UUID,
		LayoutAttr: global.INNER_LAYOUT_1,
		ComboId:    agent.UUID,
		Network:    allconnections,
		Metrics:    *utils.HostToMap(agent.Host_2, &agent.AddrInterfaceMap_2),
	}

	host_node.Tags = append(host_node.Tags, host_node.UUID, host_node.Type)
	if err := TagInjection(host_node, tagrules); err != nil {
		atomic.AddInt32(&d.agent_node_count, int32(1))
		return errors.Wrap(err, "**3")
	}

	nodes.Add(host_node)

	for _, rules := range noderules {
		uuid := ""
		for _, condition := range rules {
			if condition.Rule_type == mysqlmanager.FILTER_TYPE_HOST {
				if _uuid, ok := condition.Rule_condition["uuid"]; !ok {
					atomic.AddInt32(&d.agent_node_count, int32(1))
					return errors.Errorf("there is no uuid field in node host rule_condition: %+v **errstack**3", condition.Rule_condition)
				} else {
					uuid = _uuid.(string)
					break
				}
			}
		}
		if uuid != agent.UUID {
			continue
		}

		for _, condition := range rules {
			switch condition.Rule_type {
			case mysqlmanager.FILTER_TYPE_HOST:

			case mysqlmanager.FILTER_TYPE_PROCESS:
				for _, process := range agent.Processes_2 {
					if _name, ok := condition.Rule_condition["name"]; !ok {
						atomic.AddInt32(&d.agent_node_count, int32(1))
						return errors.Errorf("there is no name field in node rule_condition: %+v **errstack**3", condition.Rule_condition)
					} else if ProcessMatching(agent, process.ExeName, process.Cmdline, _name.(string)) {
						proc_node := &graph.Node{
							ID:         fmt.Sprintf("%s%s%s%s%d", agent.UUID, global.NODE_CONNECTOR, global.NODE_PROCESS, global.NODE_CONNECTOR, process.Pid),
							Name:       _name.(string),
							Type:       global.NODE_PROCESS,
							UUID:       agent.UUID,
							LayoutAttr: global.INNER_LAYOUT_2,
							ComboId:    agent.UUID,
							Network:    process.Connections,
							Metrics:    *utils.ProcessToMap(process),
						}

						proc_node.Tags = append(proc_node.Tags, proc_node.UUID, proc_node.Type)
						if err := TagInjection(proc_node, tagrules); err != nil {
							atomic.AddInt32(&d.agent_node_count, int32(1))
							return errors.Wrap(err, "**3")
						}

						nodes.Add(proc_node)

						break
					}
				}
			case mysqlmanager.FILTER_TYPE_TAG:
				for _, process := range agent.Processes_2 {
					if _tag, ok := condition.Rule_condition["tag_name"]; !ok {
						atomic.AddInt32(&d.agent_node_count, int32(1))
						return errors.Errorf("there is no tag_name field in node rule_condition: %+v **errstack**3", condition.Rule_condition)
					} else if ProcessMatching(agent, process.ExeName, process.Cmdline, _tag.(string)) {
						proc_node := &graph.Node{
							ID:         fmt.Sprintf("%s%s%s%s%d", agent.UUID, global.NODE_CONNECTOR, global.NODE_PROCESS, global.NODE_CONNECTOR, process.Pid),
							Name:       _tag.(string),
							Type:       global.NODE_PROCESS,
							UUID:       agent.UUID,
							LayoutAttr: global.INNER_LAYOUT_2,
							ComboId:    agent.UUID,
							Network:    process.Connections,
							Metrics:    *utils.ProcessToMap(process),
						}

						proc_node.Tags = append(proc_node.Tags, proc_node.UUID, proc_node.Type)
						if err := TagInjection(proc_node, tagrules); err != nil {
							atomic.AddInt32(&d.agent_node_count, int32(1))
							return errors.Wrap(err, "**3")
						}

						nodes.Add(proc_node)

						break
					}
				}
			case mysqlmanager.FILTER_TYPE_RESOURCE:
				for _, disk := range agent.Disks_2 {
					disk_node := &graph.Node{
						ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_RESOURCE, global.NODE_CONNECTOR, disk.Partition.Device),
						Name:       disk.Partition.Device,
						Type:       global.NODE_RESOURCE,
						UUID:       agent.UUID,
						LayoutAttr: global.INNER_LAYOUT_4,
						ComboId:    agent.UUID,
						Metrics:    *utils.DiskToMap(disk),
					}

					disk_node.Tags = append(disk_node.Tags, disk_node.UUID, disk_node.Type)
					if err := TagInjection(disk_node, tagrules); err != nil {
						atomic.AddInt32(&d.agent_node_count, int32(1))
						return errors.Wrap(err, "")
					}

					nodes.Add(disk_node)
				}

				for _, cpu := range agent.Cpus_2 {
					cpu_node := &graph.Node{
						ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_RESOURCE, global.NODE_CONNECTOR, "CPU"+strconv.Itoa(int(cpu.Info.CPU))),
						Name:       "CPU" + strconv.Itoa(int(cpu.Info.CPU)),
						Type:       global.NODE_RESOURCE,
						UUID:       agent.UUID,
						LayoutAttr: global.INNER_LAYOUT_4,
						ComboId:    agent.UUID,
						Metrics:    *utils.CpuToMap(cpu),
					}

					cpu_node.Tags = append(cpu_node.Tags, cpu_node.UUID, cpu_node.Type)
					if err := TagInjection(cpu_node, tagrules); err != nil {
						atomic.AddInt32(&d.agent_node_count, int32(1))
						return errors.Wrap(err, "**3")
					}

					nodes.Add(cpu_node)
				}

				for _, ifaceio := range agent.NetIOcounters_2 {
					iface_node := &graph.Node{
						ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_RESOURCE, global.NODE_CONNECTOR, "NC"+ifaceio.Name),
						Name:       "NC" + ifaceio.Name,
						Type:       global.NODE_RESOURCE,
						UUID:       agent.UUID,
						LayoutAttr: global.INNER_LAYOUT_4,
						ComboId:    agent.UUID,
						Metrics:    *utils.InterfaceToMap(ifaceio),
					}

					iface_node.Tags = append(iface_node.Tags, iface_node.UUID, iface_node.Type)
					if err := TagInjection(iface_node, tagrules); err != nil {
						atomic.AddInt32(&d.agent_node_count, int32(1))
						return errors.Wrap(err, "**3")
					}

					nodes.Add(iface_node)
				}
			}
		}

	}

	atomic.AddInt32(&d.agent_node_count, int32(1))

	return nil
}

func (d *DataProcesser) CreateEdgeEntities(agent *agentmanager.Agent, edges *graph.Edges, nodes *graph.Nodes) error {
	nodes_map := map[string][]*graph.Node{}

	for _, node := range nodes.Nodes {
		switch node.Type {
		case global.NODE_HOST:
			nodes_map[global.NODE_HOST] = append(nodes_map[global.NODE_HOST], node)
		case global.NODE_PROCESS:
			nodes_map[global.NODE_PROCESS] = append(nodes_map[global.NODE_PROCESS], node)
		case global.NODE_THREAD:
			nodes_map[global.NODE_THREAD] = append(nodes_map[global.NODE_THREAD], node)
		case global.NODE_NET:
			nodes_map[global.NODE_NET] = append(nodes_map[global.NODE_NET], node)
		case global.NODE_RESOURCE:
			nodes_map[global.NODE_RESOURCE] = append(nodes_map[global.NODE_RESOURCE], node)
		}
	}

	for _, sub := range nodes_map[global.NODE_HOST] {
		for _, obj := range nodes_map[global.NODE_PROCESS] {
			if obj.UUID == sub.UUID && obj.Metrics["Pid"] == "1" {
				belong_edge := &graph.Edge{
					ID:   fmt.Sprintf("%s%s%s%s%s", obj.ID, global.EDGE_CONNECTOR, global.EDGE_BELONG, global.EDGE_CONNECTOR, sub.ID),
					Type: global.EDGE_BELONG,
					Src:  obj.ID,
					Dst:  sub.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	for _, sub := range nodes_map[global.NODE_HOST] {
		for _, obj := range nodes_map[global.NODE_RESOURCE] {
			if sub.UUID == obj.UUID {
				belong_edge := &graph.Edge{
					ID:   fmt.Sprintf("%s%s%s%s%s", obj.ID, global.EDGE_CONNECTOR, global.EDGE_BELONG, global.EDGE_CONNECTOR, sub.ID),
					Type: global.EDGE_BELONG,
					Src:  obj.ID,
					Dst:  sub.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	for _, sub := range nodes_map[global.NODE_PROCESS] {
		for _, obj := range nodes_map[global.NODE_PROCESS] {
			if obj.UUID == sub.UUID && obj.Metrics["Pid"] == sub.Metrics["Ppid"] {
				belong_edge := &graph.Edge{
					ID:   fmt.Sprintf("%s%s%s%s%s", sub.ID, global.EDGE_CONNECTOR, global.EDGE_BELONG, global.EDGE_CONNECTOR, obj.ID),
					Type: global.EDGE_BELONG,
					Src:  sub.ID,
					Dst:  obj.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	// TODO: 暂定net节点关系的type均为server，暂时无法判断socket连接中的server端和agent端，需要借助其他网络工具
	for _, sub := range nodes_map[global.NODE_NET] {
		for _, obj := range nodes_map[global.NODE_PROCESS] {
			if obj.Metrics["Pid"] == sub.Metrics["Pid"] {
				server_edge := &graph.Edge{
					ID:   fmt.Sprintf("%s%s%s%s%s", sub.ID, global.EDGE_CONNECTOR, global.EDGE_SERVER, global.EDGE_CONNECTOR, obj.ID),
					Type: global.EDGE_SERVER,
					Src:  sub.ID,
					Dst:  obj.ID,
					Dir:  "direct",
				}

				edges.Add(server_edge)
			}
		}
	}

	// 生成跨主机对等网络关系实例
	for _, net := range agent.Netconnections_2 {
		var peernode1 *graph.Node
		var peernode2 *graph.Node

		for _, netnode := range nodes_map[global.NODE_NET] {
			switch netnode.Metrics["Laddr"] {
			case net.Laddr:
				peernode1 = netnode
			case net.Raddr:
				peernode2 = netnode
			}

			if peernode1 != nil && peernode2 != nil {
				break
			}
		}

		if peernode1 != nil && peernode2 != nil {
			var edgetype string
			switch peernode1.Metrics["Type"] {
			case "1":
				edgetype = global.EDGE_TCP
			case "2":
				edgetype = global.EDGE_UDP
			}

			peernet_edge := &graph.Edge{
				ID:   fmt.Sprintf("%s%s%s%s%s", peernode1.ID, global.EDGE_CONNECTOR, edgetype, global.EDGE_CONNECTOR, peernode2.ID),
				Type: edgetype,
				Src:  peernode1.ID,
				Dst:  peernode2.ID,
				Dir:  "undirect",
			}

			edges.Add(peernet_edge)
		}
	}

	return nil
}

func (d *DataProcesser) CustomCreateEdgeEntities(agent *agentmanager.Agent, edges *graph.Edges, nodes *graph.Nodes) error {
	nodes_map := map[string][]*graph.Node{}

	for _, node := range nodes.Nodes {
		switch node.Type {
		case global.NODE_HOST:
			nodes_map[global.NODE_HOST] = append(nodes_map[global.NODE_HOST], node)
		case global.NODE_PROCESS:
			nodes_map[global.NODE_PROCESS] = append(nodes_map[global.NODE_PROCESS], node)
		case global.NODE_THREAD:
			nodes_map[global.NODE_THREAD] = append(nodes_map[global.NODE_THREAD], node)
		case global.NODE_NET:
			nodes_map[global.NODE_NET] = append(nodes_map[global.NODE_NET], node)
		case global.NODE_RESOURCE:
			nodes_map[global.NODE_RESOURCE] = append(nodes_map[global.NODE_RESOURCE], node)
		}
	}

	for _, sub := range nodes_map[global.NODE_HOST] {
		for _, obj := range nodes_map[global.NODE_PROCESS] {
			if obj.UUID == sub.UUID { // && obj.Metrics["Pid"] == "1"
				belong_edge := &graph.Edge{
					ID:   fmt.Sprintf("%s%s%s%s%s", obj.ID, global.EDGE_CONNECTOR, global.EDGE_BELONG, global.EDGE_CONNECTOR, sub.ID),
					Type: global.EDGE_BELONG,
					Src:  obj.ID,
					Dst:  sub.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	// for _, sub := range nodes_map[global.NODE_HOST] {
	// 	for _, obj := range nodes_map[global.NODE_RESOURCE] {
	// 		if sub.UUID == obj.UUID {
	// 			belong_edge := &graph.Edge{
	// 				ID:   fmt.Sprintf("%s%s%s%s%s", obj.ID, global.EDGE_CONNECTOR, global.EDGE_BELONG, global.EDGE_CONNECTOR, sub.ID),
	// 				Type: global.EDGE_BELONG,
	// 				Src:  obj.ID,
	// 				Dst:  sub.ID,
	// 				Dir:  "direct",
	// 			}

	// 			edges.Add(belong_edge)
	// 		}
	// 	}
	// }

	for _, sub := range nodes_map[global.NODE_PROCESS] {
		for _, obj := range nodes_map[global.NODE_PROCESS] {
			if obj.UUID == sub.UUID && obj.Metrics["Pid"] == sub.Metrics["Ppid"] {
				belong_edge := &graph.Edge{
					ID:   fmt.Sprintf("%s%s%s%s%s", sub.ID, global.EDGE_CONNECTOR, global.EDGE_BELONG, global.EDGE_CONNECTOR, obj.ID),
					Type: global.EDGE_BELONG,
					Src:  sub.ID,
					Dst:  obj.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	// TODO: 生成对等网络关系实例, 暂时只考虑同一网段内的连接
	for _, global_net := range agent.Netconnections_2 {
		var peernode1 *graph.Node
		var peernode2 *graph.Node
		var net1 *graph.Netconnection
		var net2 *graph.Netconnection

		for _, procn := range nodes_map[global.NODE_PROCESS] {
			// 全局连接local端
			if agent.UUID == procn.UUID {
				if strconv.Itoa(int(global_net.Pid)) == procn.Metrics["Pid"] {
					peernode1 = procn
					for _, proc_net := range procn.Network {
						if proc_net.Laddr == global_net.Laddr && proc_net.Raddr == global_net.Raddr {
							net1 = &proc_net
							break
						}
					}
				}
			}
			// 全局连接remote端，同网段
			for _, proc_net := range procn.Network {
				if proc_net.Laddr == global_net.Raddr && proc_net.Raddr == global_net.Laddr {
					peernode2 = procn
					net2 = &proc_net
					break
				}
			}
			// 全局连接remote端，跨网段，本机进程为client端

			// 全局连接remote端，跨网段，本机进程为server端

			if peernode1 != nil && peernode2 != nil {
				break
			}
		}

		if peernode1 != nil && peernode2 != nil && net1 != nil && net2 != nil {
			var edgetype string
			switch global_net.Type {
			case 1:
				edgetype = global.EDGE_TCP
			case 2:
				edgetype = global.EDGE_UDP
			}

			peernet_edge := &graph.Edge{
				ID:       fmt.Sprintf("%s%s%s%s%s%s%s", peernode1.ID, global.EDGE_CONNECTOR, strings.Split(net1.Laddr, ":")[1], edgetype, strings.Split(net1.Raddr, ":")[1], global.EDGE_CONNECTOR, peernode2.ID),
				Type:     edgetype,
				Src:      peernode1.ID,
				Dst:      peernode2.ID,
				Dir:      "undirect",
				Unixtime: peernode1.Unixtime,
				Metrics: map[string]string{
					"family":    strconv.Itoa(int(net1.Family)),
					"type":      strconv.Itoa(int(net1.Type)),
					"laddr_src": net1.Laddr,
					"raddr_src": net1.Raddr,
					"laddr_dst": net2.Laddr,
					"raddr_dst": net2.Raddr,
					"status":    net1.Status,
				},
			}

			edges.Add(peernet_edge)
		}
	}

	return nil
}
