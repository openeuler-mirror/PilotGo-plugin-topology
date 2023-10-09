package processor

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/collector"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"github.com/pkg/errors"
)

type DataProcesser struct{}

var agent_node_count int
var agent_node_count_rwlock *sync.RWMutex

func CreateDataProcesser() *DataProcesser {
	return &DataProcesser{}
}

func (d *DataProcesser) Process_data() (*meta.Nodes, *meta.Edges, []error, []error) {
	nodes := &meta.Nodes{
		Lookup: make(map[string]*meta.Node, 0),
		Nodes:  make([]*meta.Node, 0),
	}
	edges := &meta.Edges{
		SrcToDsts: make(map[string][]string, 0),
		DstToSrcs: make(map[string][]string, 0),
		Edges:     make([]*meta.Edge, 0),
	}

	var wg sync.WaitGroup
	agent_count := 0
	create_node_rwlock := &sync.RWMutex{}
	agent_node_count = 0
	agent_node_count_rwlock = &sync.RWMutex{}
	var collect_errorlist []error
	var process_errorlist []error

	// 获取运行状态agent的数目
	agentmanager.Topo.AgentMap.Range(func(key, value interface{}) bool {
		agent := value.(*agentmanager.Agent_m)
		if agent.Host_2 != nil {
			agent_count++
		}

		return true
	})

	datacollector := collector.CreateDataCollector()
	collect_errorlist = datacollector.Collect_instant_data()
	if len(collect_errorlist) != 0 {
		for i, err := range collect_errorlist {
			collect_errorlist[i] = errors.Wrap(err, "**7")
		}

		// return nil, nil, collect_errorlist, nil
	}

	agentmanager.Topo.AgentMap.Range(
		func(key, value interface{}) bool {
			wg.Add(1)

			go func() {
				defer wg.Done()
				agent := value.(*agentmanager.Agent_m)

				if agent.Host_2 != nil && agent.Processes_2 != nil && agent.Netconnections_2 != nil {
					err := d.Create_node_entities(agent, nodes, create_node_rwlock)
					if err != nil {
						process_errorlist = append(process_errorlist, errors.Wrap(err, "**2"))
					}

					for {
						if agent_node_count == agent_count {
							break
						}
						// ttcode
						// fmt.Printf("\033[32m agent_node_count\033[0m: %d\n", agent_node_count)
						// fmt.Printf("\033[32magent_count\033[0m: %d\n", agent_count)
					}

					err = d.Create_edge_entities(agent, edges, nodes)
					if err != nil {
						process_errorlist = append(process_errorlist, errors.Wrap(err, "**2"))
					}
				}
			}()

			return true
		},
	)
	wg.Wait()

	return nodes, edges, collect_errorlist, process_errorlist
}

func (d *DataProcesser) Create_node_entities(agent *agentmanager.Agent_m, nodes *meta.Nodes, mu *sync.RWMutex) error {
	host_node := &meta.Node{
		ID:      fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_HOST, agent.IP),
		Name:    agent.UUID,
		Type:    meta.NODE_HOST,
		UUID:    agent.UUID,
		Metrics: *utils.HostToMap(agent.Host_2, &agent.AddrInterfaceMap_2),
	}

	mu.Lock()
	nodes.Add(host_node)
	mu.Unlock()

	for _, process := range agent.Processes_2 {
		proc_node := &meta.Node{
			ID:      fmt.Sprintf("%s_%s_%d", agent.UUID, meta.NODE_PROCESS, process.Pid),
			Name:    process.ExeName,
			Type:    meta.NODE_PROCESS,
			UUID:    agent.UUID,
			Metrics: *utils.ProcessToMap(process),
		}

		mu.Lock()
		nodes.Add(proc_node)
		mu.Unlock()

		for _, thread := range process.Threads {
			thre_node := &meta.Node{
				ID:      fmt.Sprintf("%s_%s_%d", agent.UUID, meta.NODE_THREAD, thread.Tid),
				Name:    strconv.Itoa(int(thread.Tid)),
				Type:    meta.NODE_THREAD,
				UUID:    agent.UUID,
				Metrics: *utils.ThreadToMap(&thread),
			}

			mu.Lock()
			nodes.Add(thre_node)
			mu.Unlock()
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
		net_node := &meta.Node{
			ID:      fmt.Sprintf("%s_%s_%d:%s", agent.UUID, meta.NODE_NET, net.Pid, strings.Split(net.Laddr, ":")[1]),
			Name:    net.Laddr,
			Type:    meta.NODE_NET,
			UUID:    agent.UUID,
			Metrics: *utils.NetToMap(net),
		}

		mu.Lock()
		nodes.Add(net_node)
		mu.Unlock()
	}

	for _, disk := range agent.Disks_2 {
		disk_node := &meta.Node{
			ID:      fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, disk.Partition.Device),
			Name:    disk.Partition.Device,
			Type:    meta.NODE_RESOURCE,
			UUID:    agent.UUID,
			Metrics: *utils.DiskToMap(disk),
		}

		mu.Lock()
		nodes.Add(disk_node)
		mu.Unlock()
	}

	for _, cpu := range agent.Cpus_2 {
		cpu_node := &meta.Node{
			ID:      fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, "CPU"+strconv.Itoa(int(cpu.Info.CPU))),
			Name:    "CPU" + strconv.Itoa(int(cpu.Info.CPU)),
			Type:    meta.NODE_RESOURCE,
			UUID:    agent.UUID,
			Metrics: *utils.CpuToMap(cpu),
		}

		mu.Lock()
		nodes.Add(cpu_node)
		mu.Unlock()
	}

	for _, ifaceio := range agent.NetIOcounters_2 {
		iface_node := &meta.Node{
			ID:      fmt.Sprintf("%s_%s_%s", agent.UUID, meta.NODE_RESOURCE, "NC"+ifaceio.Name),
			Name:    "NC" + ifaceio.Name,
			Type:    meta.NODE_RESOURCE,
			UUID:    agent.UUID,
			Metrics: *utils.InterfaceToMap(ifaceio),
		}

		mu.Lock()
		nodes.Add(iface_node)
		mu.Unlock()
	}

	agent_node_count_rwlock.Lock()
	agent_node_count++
	agent_node_count_rwlock.Unlock()

	return nil
}

func (d *DataProcesser) Create_edge_entities(agent *agentmanager.Agent_m, edges *meta.Edges, nodes *meta.Nodes) error {
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

	// TODO: edge实例重复
	for _, obj := range nodes_map[meta.NODE_HOST] {
		for _, sub := range nodes_map[meta.NODE_PROCESS] {
			if sub.Metrics["Pid"] == "1" && sub.UUID == obj.UUID {
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", sub.ID, meta.EDGE_BELONG, obj.ID),
					Type: meta.EDGE_BELONG,
					Src:  sub.ID,
					Dst:  obj.ID,
					Dir:  true,
				}

				edges.Add(belong_edge)
			}
		}
	}

	for _, obj := range nodes_map[meta.NODE_HOST] {
		for _, sub := range nodes_map[meta.NODE_RESOURCE] {
			if sub.UUID == obj.UUID {
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", sub.ID, meta.EDGE_BELONG, obj.ID),
					Type: meta.EDGE_BELONG,
					Src:  sub.ID,
					Dst:  obj.ID,
					Dir:  true,
				}

				edges.Add(belong_edge)
			}
		}
	}

	for _, sub := range nodes_map[meta.NODE_PROCESS] {
		for _, obj := range nodes_map[meta.NODE_PROCESS] {
			if obj.Metrics["Pid"] == sub.Metrics["Ppid"] && obj.UUID == sub.UUID {
				belong_edge := &meta.Edge{
					ID:   fmt.Sprintf("%s_%s_%s", sub.ID, meta.EDGE_BELONG, obj.ID),
					Type: meta.EDGE_BELONG,
					Src:  sub.ID,
					Dst:  obj.ID,
					Dir:  true,
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
					Dir:  true,
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
				Dir:  false,
			}

			edges.Add(peernet_edge)
		}
	}

	return nil
}
