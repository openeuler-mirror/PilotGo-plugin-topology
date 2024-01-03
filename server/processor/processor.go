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
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
)

type DataProcesser struct{}

var agent_node_count int32

func CreateDataProcesser() *DataProcesser {
	return &DataProcesser{}
}

func (d *DataProcesser) Process_data(agentnum int) (*meta.Nodes, *meta.Edges, []error, []error) {
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

	datacollector := collector.CreateDataCollector()
	collect_errorlist = datacollector.Collect_instant_data()
	if len(collect_errorlist) != 0 {
		for i, err := range collect_errorlist {
			collect_errorlist[i] = errors.Wrap(err, "**7")
		}

		// return nil, nil, collect_errorlist, nil
	}

	ctx1, cancel1 := context.WithCancel(agentmanager.Topo.Tctx)
	go func(cancelfunc context.CancelFunc) {
		for {
			if atomic.LoadInt32(&agent_node_count) == int32(agentnum) {
				cancelfunc()
				break
			}
		}
	}(cancel1)

	agentmanager.Topo.TAgentMap.Range(
		func(key, value interface{}) bool {
			wg.Add(1)

			agent := value.(*agentmanager.Agent_m)

			go func(ctx context.Context, _agent *agentmanager.Agent_m, _nodes *meta.Nodes, _edges *meta.Edges) {
				defer wg.Done()

				if _agent.Host_2 != nil && _agent.Processes_2 != nil && _agent.Netconnections_2 != nil {
					err := d.Create_node_entities(_agent, _nodes)
					if err != nil {
						process_errorlist = append(process_errorlist, errors.Wrap(err, "**2"))
					}

					<-ctx.Done()

					err = d.Create_edge_entities(_agent, _edges, _nodes)
					if err != nil {
						process_errorlist = append(process_errorlist, errors.Wrap(err, "**2"))
					}
				}
			}(ctx1, agent, nodes, edges)

			return true
		},
	)
	wg.Wait()

	atomic.StoreInt32(&agent_node_count, int32(0))

	elapse := time.Since(start)
	// fmt.Fprintf(agentmanager.Topo.Out, "\033[32mtopo server 采集数据处理时间\033[0m: %v\n", elapse)
	logger.Info("\033[32mtopo server 采集数据处理时间\033[0m: %v\n", elapse)

	return nodes, edges, collect_errorlist, process_errorlist
}

func (d *DataProcesser) Create_node_entities(agent *agentmanager.Agent_m, nodes *meta.Nodes) error {
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
		net_node := &meta.Node{
			ID:         fmt.Sprintf("%s_%s_%d:%s", agent.UUID, meta.NODE_NET, net.Pid, strings.Split(net.Laddr, ":")[1]),
			Name:       net.Laddr,
			Type:       meta.NODE_NET,
			UUID:       agent.UUID,
			LayoutAttr: "d",
			ComboId:    agent.UUID,
			Metrics:    *utils.NetToMap(net),
		}

		nodes.Add(net_node)
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

	atomic.AddInt32(&agent_node_count, int32(1))

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

	for _, obj := range nodes_map[meta.NODE_HOST] {
		for _, sub := range nodes_map[meta.NODE_PROCESS] {
			if sub.UUID == obj.UUID && sub.Metrics["Pid"] == "1" {
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

	for _, obj := range nodes_map[meta.NODE_HOST] {
		for _, sub := range nodes_map[meta.NODE_RESOURCE] {
			if sub.UUID == obj.UUID {
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

	for _, sub := range nodes_map[meta.NODE_PROCESS] {
		for _, obj := range nodes_map[meta.NODE_PROCESS] {
			if obj.Metrics["Pid"] == sub.Metrics["Ppid"] && obj.UUID == sub.UUID {
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
