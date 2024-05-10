package generator

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/graph"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"github.com/pkg/errors"
)

type PublicTopo struct {
	Agent_node_count *int32
}

func (p *PublicTopo) CreateNodeEntities(agent *agentmanager.Agent, nodes *graph.Nodes) error {
	host_node := &graph.Node{
		ID:         fmt.Sprintf("%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_HOST, global.NODE_CONNECTOR, agent.IP),
		Name:       agent.UUID,
		Type:       global.NODE_HOST,
		UUID:       agent.UUID,
		LayoutAttr: global.INNER_LAYOUT_1,
		ComboId:    agent.UUID,
		Metrics:    *graph.HostToMap(agent.Host_2, &agent.AddrInterfaceMap_2),
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
			Metrics:    *graph.ProcessToMap(process),
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
				Metrics:    *graph.ThreadToMap(&thread),
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
				Metrics:    *graph.NetToMap(net),
			}

			nodes.Add(net_node)
		} else {
			err := errors.Errorf("syntax error: %s **errstack**13", net.Laddr) // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
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
			Metrics:    *graph.DiskToMap(disk),
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
			Metrics:    *graph.CpuToMap(cpu),
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
			Metrics:    *graph.InterfaceToMap(ifaceio),
		}

		nodes.Add(iface_node)
	}

	atomic.AddInt32(p.Agent_node_count, int32(1))

	return nil
}

func (p *PublicTopo) CreateEdgeEntities(agent *agentmanager.Agent, edges *graph.Edges, nodes *graph.Nodes) error {
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

func (p *PublicTopo) Return_Agent_node_count() *int32 {
	return p.Agent_node_count
}
