package processor

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/collector"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
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

func (d *DataProcesser) ProcessData(agentnum int, tagrules []meta.Tag_rule, noderules [][]meta.Filter_rule) (*meta.Nodes, *meta.Edges, []error, []error) {
	start := time.Now()
	nodes := &meta.Nodes{
		Lock:         sync.Mutex{},
		Lookup:       make(map[string]*meta.Node, 0),
		LookupByType: make(map[string][]*meta.Node, 0),
		LookupByUUID: make(map[string][]*meta.Node, 0),
		Nodes:        make([]*meta.Node, 0),
	}
	edges := &meta.Edges{
		Lock:      sync.Mutex{},
		Lookup:    sync.Map{},
		SrcToDsts: make(map[string][]string, 0),
		DstToSrcs: make(map[string][]string, 0),
		Edges:     make([]*meta.Edge, 0),
	}

	var wg sync.WaitGroup
	var collect_errorlist []error
	var process_errorlist []error
	var process_errorlist_rwlock sync.RWMutex

	if agentmanager.Topo == nil {
		err := errors.New("agentmanager.Topo is not initialized!") // err top
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	datacollector := collector.CreateDataCollector()
	collect_errorlist = datacollector.CollectInstantData()
	if len(collect_errorlist) != 0 {
		for i, err := range collect_errorlist {
			collect_errorlist[i] = errors.Wrap(err, "**7")
		}
	}

	ctx1, cancel1 := context.WithCancel(agentmanager.Topo.Tctx)
	go func(cancelfunc context.CancelFunc) {
		for {
			if atomic.LoadInt32(&d.agent_node_count) == int32(agentnum) {
				cancelfunc()
				break
			}
		}
	}(cancel1)

	agentmanager.Topo.TAgentMap.Range(
		func(key, value interface{}) bool {
			wg.Add(1)

			agent := value.(*agentmanager.Agent_m)

			go func(ctx context.Context, _agent *agentmanager.Agent_m, _nodes *meta.Nodes, _edges *meta.Edges, _tagrules []meta.Tag_rule, _noderules [][]meta.Filter_rule) {
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

func (d *DataProcesser) CreateNodeEntities(agent *agentmanager.Agent_m, nodes *meta.Nodes) error {
	host_node := &meta.Node{
		ID:         fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_HOST, agent.IP),
		Name:       agent.UUID,
		Type:       meta.NODE_HOST,
		UUID:       agent.UUID,
		LayoutAttr: "a",
		ComboId:    agent.UUID,
		Metrics:    *utils.HostToMap(agent.Host_2, &agent.AddrInterfaceMap_2),
	}

	nodes.Add(host_node)

	for _, process := range agent.Processes_2 {
		proc_node := &meta.Node{
			ID:         fmt.Sprintf("%s_%s_%d", agent.UUID, meta.NODE_PROCESS, process.Pid),
			Name:       process.ExeName,
			Type:       meta.NODE_PROCESS,
			UUID:       agent.UUID,
			LayoutAttr: "b",
			ComboId:    agent.UUID,
			Metrics:    *utils.ProcessToMap(process),
		}

		nodes.Add(proc_node)

		for _, thread := range process.Threads {
			thre_node := &meta.Node{
				ID:         fmt.Sprintf("%s_%s_%d", agent.UUID, meta.NODE_THREAD, thread.Tid),
				Name:       strconv.Itoa(int(thread.Tid)),
				Type:       meta.NODE_THREAD,
				UUID:       agent.UUID,
				LayoutAttr: "c",
				ComboId:    agent.UUID,
				Metrics:    *utils.ThreadToMap(&thread),
			}

			nodes.Add(thre_node)
		}

		// for _, net := range process.NetIOCounters {
		// 	net_node := &meta.Node{
		// 		ID:      fmt.Sprintf("%s-%s-%d", agent.UUID, meta.NODE_NET, process.Pid),
		// 		Name:    net.Name,
		// 		Type:    meta.NODE_NET,
		// 		UUID:    agent.UUID,
		// 		Metrics: *utils.NetToMap(&net, &agent.AddrInterfaceMap_2),
		// 	}

		// 	nodes.Add(net_node)
		// }
	}

	// 临时定义不含网络流量metric的网络节点
	for _, net := range agent.Netconnections_2 {
		if laddr_slice := strings.Split(net.Laddr, ":"); len(laddr_slice) != 0 {
			net_node := &meta.Node{
				ID:         fmt.Sprintf("%s_%s_%d:%s", agent.UUID, meta.NODE_NET, net.Pid, laddr_slice[1]),
				Name:       net.Laddr,
				Type:       meta.NODE_NET,
				UUID:       agent.UUID,
				LayoutAttr: "d",
				ComboId:    agent.UUID,
				Metrics:    *utils.NetToMap(net),
			}

			nodes.Add(net_node)
		} else {
			err := errors.Errorf("syntax error: %s **warn**13", net.Laddr) // err top
			agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		}
	}

	for _, disk := range agent.Disks_2 {
		disk_node := &meta.Node{
			ID:         fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, disk.Partition.Device),
			Name:       disk.Partition.Device,
			Type:       meta.NODE_RESOURCE,
			UUID:       agent.UUID,
			LayoutAttr: "e",
			ComboId:    agent.UUID,
			Metrics:    *utils.DiskToMap(disk),
		}

		nodes.Add(disk_node)
	}

	for _, cpu := range agent.Cpus_2 {
		cpu_node := &meta.Node{
			ID:         fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, "CPU"+strconv.Itoa(int(cpu.Info.CPU))),
			Name:       "CPU" + strconv.Itoa(int(cpu.Info.CPU)),
			Type:       meta.NODE_RESOURCE,
			UUID:       agent.UUID,
			LayoutAttr: "e",
			ComboId:    agent.UUID,
			Metrics:    *utils.CpuToMap(cpu),
		}

		nodes.Add(cpu_node)
	}

	for _, ifaceio := range agent.NetIOcounters_2 {
		iface_node := &meta.Node{
			ID:         fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, "NC"+ifaceio.Name),
			Name:       "NC" + ifaceio.Name,
			Type:       meta.NODE_RESOURCE,
			UUID:       agent.UUID,
			LayoutAttr: "e",
			ComboId:    agent.UUID,
			Metrics:    *utils.InterfaceToMap(ifaceio),
		}

		nodes.Add(iface_node)
	}

	atomic.AddInt32(&d.agent_node_count, int32(1))

	return nil
}

func (d *DataProcesser) CustomCreateNodeEntities(agent *agentmanager.Agent_m, nodes *meta.Nodes, tagrules []meta.Tag_rule, noderules [][]meta.Filter_rule) error {
	host_node := &meta.Node{
		ID:         fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_HOST, agent.IP),
		Name:       agent.UUID,
		Type:       meta.NODE_HOST,
		UUID:       agent.UUID,
		LayoutAttr: "a",
		ComboId:    agent.UUID,
		Metrics:    *utils.HostToMap(agent.Host_2, &agent.AddrInterfaceMap_2),
	}

	host_node.Tags = append(host_node.Tags, host_node.UUID, host_node.Type)
	host_node = utils.TagInjection(host_node, tagrules)

	nodes.Add(host_node)

	for _, rules := range noderules {
		uuid := ""
		for _, condition := range rules {
			if condition.Rule_type == meta.FILTER_TYPE_HOST {
				uuid = condition.Rule_condition["uuid"]
				break
			}	
		}
		if uuid != agent.UUID {
			continue
		}

		for _, condition := range rules {
			switch condition.Rule_type {
			case meta.FILTER_TYPE_HOST:

			case meta.FILTER_TYPE_PROCESS:
				for _, process := range agent.Processes_2 {
					if condition.Rule_condition["name"] == process.ExeName {
						proc_node := &meta.Node{
							ID:         fmt.Sprintf("%s_%s_%d", agent.UUID, meta.NODE_PROCESS, process.Pid),
							Name:       process.ExeName,
							Type:       meta.NODE_PROCESS,
							UUID:       agent.UUID,
							LayoutAttr: "b",
							ComboId:    agent.UUID,
							Metrics:    *utils.ProcessToMap(process),
						}

						proc_node.Tags = append(proc_node.Tags, proc_node.UUID, proc_node.Type)
						proc_node = utils.TagInjection(proc_node, tagrules)

						nodes.Add(proc_node)

						break
					}
				}
			case meta.FILTER_TYPE_TAG:
				for _, process := range agent.Processes_2 {
					if condition.Rule_condition["tag_name"] == process.ExeName {
						proc_node := &meta.Node{
							ID:         fmt.Sprintf("%s_%s_%d", agent.UUID, meta.NODE_PROCESS, process.Pid),
							Name:       process.ExeName,
							Type:       meta.NODE_PROCESS,
							UUID:       agent.UUID,
							LayoutAttr: "b",
							ComboId:    agent.UUID,
							Metrics:    *utils.ProcessToMap(process),
						}

						proc_node.Tags = append(proc_node.Tags, proc_node.UUID, proc_node.Type)
						proc_node = utils.TagInjection(proc_node, tagrules)

						nodes.Add(proc_node)

						break
					}
				}
			case meta.FILTER_TYPE_RESOURCE:
				for _, disk := range agent.Disks_2 {
					disk_node := &meta.Node{
						ID:         fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, disk.Partition.Device),
						Name:       disk.Partition.Device,
						Type:       meta.NODE_RESOURCE,
						UUID:       agent.UUID,
						LayoutAttr: "e",
						ComboId:    agent.UUID,
						Metrics:    *utils.DiskToMap(disk),
					}

					disk_node.Tags = append(disk_node.Tags, disk_node.UUID, disk_node.Type)
					disk_node = utils.TagInjection(disk_node, tagrules)

					nodes.Add(disk_node)
				}

				for _, cpu := range agent.Cpus_2 {
					cpu_node := &meta.Node{
						ID:         fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, "CPU"+strconv.Itoa(int(cpu.Info.CPU))),
						Name:       "CPU" + strconv.Itoa(int(cpu.Info.CPU)),
						Type:       meta.NODE_RESOURCE,
						UUID:       agent.UUID,
						LayoutAttr: "e",
						ComboId:    agent.UUID,
						Metrics:    *utils.CpuToMap(cpu),
					}

					cpu_node.Tags = append(cpu_node.Tags, cpu_node.UUID, cpu_node.Type)
					cpu_node = utils.TagInjection(cpu_node, tagrules)

					nodes.Add(cpu_node)
				}

				for _, ifaceio := range agent.NetIOcounters_2 {
					iface_node := &meta.Node{
						ID:         fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, "NC"+ifaceio.Name),
						Name:       "NC" + ifaceio.Name,
						Type:       meta.NODE_RESOURCE,
						UUID:       agent.UUID,
						LayoutAttr: "e",
						ComboId:    agent.UUID,
						Metrics:    *utils.InterfaceToMap(ifaceio),
					}

					iface_node.Tags = append(iface_node.Tags, iface_node.UUID, iface_node.Type)
					iface_node = utils.TagInjection(iface_node, tagrules)

					nodes.Add(iface_node)
				}
			}
		}

	}

	atomic.AddInt32(&d.agent_node_count, int32(1))

	return nil
}

func (d *DataProcesser) CreateEdgeEntities(agent *agentmanager.Agent_m, edges *meta.Edges, nodes *meta.Nodes) error {
	nodes_map := map[string][]*meta.Node{}

	for _, node := range nodes.Nodes {
		switch node.Type {
		case meta.NODE_HOST:
			nodes_map[meta.NODE_HOST] = append(nodes_map[meta.NODE_HOST], node)
		case meta.NODE_PROCESS:
			nodes_map[meta.NODE_PROCESS] = append(nodes_map[meta.NODE_PROCESS], node)
		case meta.NODE_THREAD:
			nodes_map[meta.NODE_THREAD] = append(nodes_map[meta.NODE_THREAD], node)
		case meta.NODE_NET:
			nodes_map[meta.NODE_NET] = append(nodes_map[meta.NODE_NET], node)
		case meta.NODE_RESOURCE:
			nodes_map[meta.NODE_RESOURCE] = append(nodes_map[meta.NODE_RESOURCE], node)
		}
	}

	for _, sub := range nodes_map[meta.NODE_HOST] {
		for _, obj := range nodes_map[meta.NODE_PROCESS] {
			if obj.UUID == sub.UUID && obj.Metrics["Pid"] == "1" {
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", obj.ID, meta.EDGE_BELONG, sub.ID),
					Type: meta.EDGE_BELONG,
					Src:  obj.ID,
					Dst:  sub.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	for _, sub := range nodes_map[meta.NODE_HOST] {
		for _, obj := range nodes_map[meta.NODE_RESOURCE] {
			if sub.UUID == obj.UUID {
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", obj.ID, meta.EDGE_BELONG, sub.ID),
					Type: meta.EDGE_BELONG,
					Src:  obj.ID,
					Dst:  sub.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	for _, sub := range nodes_map[meta.NODE_PROCESS] {
		for _, obj := range nodes_map[meta.NODE_PROCESS] {
			if obj.UUID == sub.UUID && obj.Metrics["Pid"] == sub.Metrics["Ppid"] {
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", sub.ID, meta.EDGE_BELONG, obj.ID),
					Type: meta.EDGE_BELONG,
					Src:  sub.ID,
					Dst:  obj.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	// TODO: 暂定net节点关系的type均为server，暂时无法判断socket连接中的server端和agent端，需要借助其他网络工具
	for _, sub := range nodes_map[meta.NODE_NET] {
		for _, obj := range nodes_map[meta.NODE_PROCESS] {
			if obj.Metrics["Pid"] == sub.Metrics["Pid"] {
				server_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", sub.ID, meta.EDGE_SERVER, obj.ID),
					Type: meta.EDGE_SERVER,
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
		var peernode1 *meta.Node
		var peernode2 *meta.Node

		for _, netnode := range nodes_map[meta.NODE_NET] {
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
				edgetype = meta.EDGE_TCP
			case "2":
				edgetype = meta.EDGE_UDP
			}

			peernet_edge := &meta.Edge{
				ID:   fmt.Sprintf("%s_%s_%s", peernode1.ID, edgetype, peernode2.ID),
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

func (d *DataProcesser) CustomCreateEdgeEntities(agent *agentmanager.Agent_m, edges *meta.Edges, nodes *meta.Nodes) error {
	nodes_map := map[string][]*meta.Node{}

	for _, node := range nodes.Nodes {
		switch node.Type {
		case meta.NODE_HOST:
			nodes_map[meta.NODE_HOST] = append(nodes_map[meta.NODE_HOST], node)
		case meta.NODE_PROCESS:
			nodes_map[meta.NODE_PROCESS] = append(nodes_map[meta.NODE_PROCESS], node)
		case meta.NODE_THREAD:
			nodes_map[meta.NODE_THREAD] = append(nodes_map[meta.NODE_THREAD], node)
		case meta.NODE_NET:
			nodes_map[meta.NODE_NET] = append(nodes_map[meta.NODE_NET], node)
		case meta.NODE_RESOURCE:
			nodes_map[meta.NODE_RESOURCE] = append(nodes_map[meta.NODE_RESOURCE], node)
		}
	}

	for _, sub := range nodes_map[meta.NODE_HOST] {
		for _, obj := range nodes_map[meta.NODE_PROCESS] {
			if obj.UUID == sub.UUID { // && obj.Metrics["Pid"] == "1"
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", obj.ID, meta.EDGE_BELONG, sub.ID),
					Type: meta.EDGE_BELONG,
					Src:  obj.ID,
					Dst:  sub.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	for _, sub := range nodes_map[meta.NODE_HOST] {
		for _, obj := range nodes_map[meta.NODE_RESOURCE] {
			if sub.UUID == obj.UUID {
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", obj.ID, meta.EDGE_BELONG, sub.ID),
					Type: meta.EDGE_BELONG,
					Src:  obj.ID,
					Dst:  sub.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	for _, sub := range nodes_map[meta.NODE_PROCESS] {
		for _, obj := range nodes_map[meta.NODE_PROCESS] {
			if obj.UUID == sub.UUID && obj.Metrics["Pid"] == sub.Metrics["Ppid"] {
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", sub.ID, meta.EDGE_BELONG, obj.ID),
					Type: meta.EDGE_BELONG,
					Src:  sub.ID,
					Dst:  obj.ID,
					Dir:  "direct",
				}

				edges.Add(belong_edge)
			}
		}
	}

	// TODO: 暂定net节点关系的type均为server，暂时无法判断socket连接中的server端和agent端，需要借助其他网络工具
	for _, sub := range nodes_map[meta.NODE_NET] {
		for _, obj := range nodes_map[meta.NODE_PROCESS] {
			if obj.Metrics["Pid"] == sub.Metrics["Pid"] {
				server_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", sub.ID, meta.EDGE_SERVER, obj.ID),
					Type: meta.EDGE_SERVER,
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
		var peernode1 *meta.Node
		var peernode2 *meta.Node

		for _, netnode := range nodes_map[meta.NODE_NET] {
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
				edgetype = meta.EDGE_TCP
			case "2":
				edgetype = meta.EDGE_UDP
			}

			peernet_edge := &meta.Edge{
				ID:   fmt.Sprintf("%s_%s_%s", peernode1.ID, edgetype, peernode2.ID),
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
