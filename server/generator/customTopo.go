package generator

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/generator/utils"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/graph"
	"github.com/pkg/errors"
)

type CustomTopo struct {
	Tagrules  []mysqlmanager.Tag_rule
	Noderules [][]mysqlmanager.Filter_rule

	Agent_node_count *int32
}

func (c *CustomTopo) CreateNodeEntities(agent *agentmanager.Agent, nodes *graph.Nodes) error {
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
		Metrics:    *graph.HostToMap(agent.Host_2, &agent.AddrInterfaceMap_2),
	}

	host_node.Tags = append(host_node.Tags, host_node.UUID, host_node.Type)
	if err := utils.TagInjection(host_node, c.Tagrules); err != nil {
		atomic.AddInt32(c.Agent_node_count, int32(1))
		return errors.Wrap(err, " ")
	}

	nodes.Add(host_node)

	for _, rules := range c.Noderules {
		uuid := ""
		for _, condition := range rules {
			if condition.Rule_type == mysqlmanager.FILTER_TYPE_HOST {
				if _uuid, ok := condition.Rule_condition["uuid"]; !ok {
					atomic.AddInt32(c.Agent_node_count, int32(1))
					return errors.Errorf("there is no uuid field in node host rule_condition: %+v", condition.Rule_condition)
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
						atomic.AddInt32(c.Agent_node_count, int32(1))
						return errors.Errorf("there is no name field in node rule_condition: %+v", condition.Rule_condition)
					} else if utils.ProcessMatching(agent, process.ExeName, process.Cmdline, _name.(string)) {
						metrics_map := *graph.ProcessToMap(process)
						proc_node := &graph.Node{
							ID:         fmt.Sprintf("%s%s%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_PROCESS, global.NODE_CONNECTOR, _name.(string), global.NODE_CONNECTOR, global.GenerateRandomID(5)),
							Name:       _name.(string),
							Type:       global.NODE_PROCESS,
							UUID:       agent.UUID,
							LayoutAttr: global.INNER_LAYOUT_2,
							ComboId:    agent.UUID,
							Network:    process.Connections,
							Metrics:    metrics_map,
						}

						proc_node.Tags = append(proc_node.Tags, proc_node.UUID, proc_node.Type)
						if err := utils.TagInjection(proc_node, c.Tagrules); err != nil {
							atomic.AddInt32(c.Agent_node_count, int32(1))
							return errors.Wrap(err, " ")
						}

						nodes.Add(proc_node)

						break
					}
				}
			case mysqlmanager.FILTER_TYPE_TAG:
				for _, process := range agent.Processes_2 {
					if _tag, ok := condition.Rule_condition["tag_name"]; !ok {
						atomic.AddInt32(c.Agent_node_count, int32(1))
						return errors.Errorf("there is no tag_name field in node rule_condition: %+v", condition.Rule_condition)
					} else if utils.ProcessMatching(agent, process.ExeName, process.Cmdline, _tag.(string)) {
						metrics_map := *graph.ProcessToMap(process)
						proc_node := &graph.Node{
							ID:         fmt.Sprintf("%s%s%s%s%s%s%s", agent.UUID, global.NODE_CONNECTOR, global.NODE_PROCESS, global.NODE_CONNECTOR,  _tag.(string), global.NODE_CONNECTOR, global.GenerateRandomID(5)),
							Name:       _tag.(string),
							Type:       global.NODE_PROCESS,
							UUID:       agent.UUID,
							LayoutAttr: global.INNER_LAYOUT_2,
							ComboId:    agent.UUID,
							Network:    process.Connections,
							Metrics:    metrics_map,
						}

						proc_node.Tags = append(proc_node.Tags, proc_node.UUID, proc_node.Type)
						if err := utils.TagInjection(proc_node, c.Tagrules); err != nil {
							atomic.AddInt32(c.Agent_node_count, int32(1))
							return errors.Wrap(err, " ")
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
						Metrics:    *graph.DiskToMap(disk),
					}

					disk_node.Tags = append(disk_node.Tags, disk_node.UUID, disk_node.Type)
					if err := utils.TagInjection(disk_node, c.Tagrules); err != nil {
						atomic.AddInt32(c.Agent_node_count, int32(1))
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
						Metrics:    *graph.CpuToMap(cpu),
					}

					cpu_node.Tags = append(cpu_node.Tags, cpu_node.UUID, cpu_node.Type)
					if err := utils.TagInjection(cpu_node, c.Tagrules); err != nil {
						atomic.AddInt32(c.Agent_node_count, int32(1))
						return errors.Wrap(err, " ")
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

					iface_node.Tags = append(iface_node.Tags, iface_node.UUID, iface_node.Type)
					if err := utils.TagInjection(iface_node, c.Tagrules); err != nil {
						atomic.AddInt32(c.Agent_node_count, int32(1))
						return errors.Wrap(err, " ")
					}

					nodes.Add(iface_node)
				}
			}
		}

	}

	atomic.AddInt32(c.Agent_node_count, int32(1))

	return nil
}

func (c *CustomTopo) CreateEdgeEntities(agent *agentmanager.Agent, edges *graph.Edges, nodes *graph.Nodes) error {
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
		var exist_multi_net bool
		var multi_net_edge *graph.Edge

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

			net_metrics := map[string]string{
				fmt.Sprintf("%s_%s_family", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1]):    strconv.Itoa(int(net1.Family)),
				fmt.Sprintf("%s_%s_type", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1]):      strconv.Itoa(int(net1.Type)),
				fmt.Sprintf("%s_%s_laddr_src", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1]): net1.Laddr,
				fmt.Sprintf("%s_%s_raddr_src", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1]): net1.Raddr,
				fmt.Sprintf("%s_%s_laddr_dst", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1]): net2.Laddr,
				fmt.Sprintf("%s_%s_raddr_dst", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1]): net2.Raddr,
				fmt.Sprintf("%s_%s_status", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1]):    net1.Status,
			}

			// 两个进程之间存在多个网络连接时，将多个网络连接放入一个边实例中
			for _, edge := range edges.Edges {
				if (edge.Src == peernode1.ID && edge.Dst == peernode2.ID) || (edge.Src == peernode2.ID && edge.Dst == peernode1.ID) {
					exist_multi_net = true
					multi_net_edge = edge
				}
			}

			if exist_multi_net {
				for _, m := range multi_net_edge.Metrics {
					if m[fmt.Sprintf("%s_%s_laddr_src", strings.Split(net1.Raddr, ":")[1], strings.Split(net1.Laddr, ":")[1])] == net_metrics[fmt.Sprintf("%s_%s_laddr_dst", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1])] && m[fmt.Sprintf("%s_%s_laddr_dst", strings.Split(net1.Raddr, ":")[1], strings.Split(net1.Laddr, ":")[1])] == net_metrics[fmt.Sprintf("%s_%s_laddr_src", strings.Split(net1.Laddr, ":")[1], strings.Split(net1.Raddr, ":")[1])] {
						goto jump
					}
				}
				multi_net_edge.Metrics = append(multi_net_edge.Metrics, net_metrics)
			jump:
			} else {
				peernet_edge := &graph.Edge{
					ID:       fmt.Sprintf("%s%s%s%s%s", peernode1.ID, global.EDGE_CONNECTOR, edgetype, global.EDGE_CONNECTOR, peernode2.ID),
					Type:     edgetype,
					Src:      peernode1.ID,
					Dst:      peernode2.ID,
					Dir:      "undirect",
					Unixtime: peernode1.Unixtime,
					Metrics:  []map[string]string{net_metrics},
				}
				edges.Add(peernet_edge)
			}
		}
	}

	return nil
}

func (c *CustomTopo) Return_Agent_node_count() *int32 {
	return c.Agent_node_count
}
