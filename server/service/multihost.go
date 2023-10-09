package service

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
	"github.com/pkg/errors"
)

func MultiHostService() ([]*meta.Node, []*meta.Edge, []error, []error) {
	dataprocesser := processor.CreateDataProcesser()
	nodes, edges, collect_errlist, process_errlist := dataprocesser.Process_data()
	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, "**3")
		}

		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, "**7")
		}
	}

	hostids := []string{}
	multi_nodes_map := make(map[string]*meta.Node)
	multi_nodes := []*meta.Node{}
	multi_edges_map := make(map[string]*meta.Edge)
	multi_edges := []*meta.Edge{}

	// 添加 host node
	for _, node := range nodes.Nodes {
		if node.Type == "host" {
			if _, ok := multi_nodes_map[node.ID]; !ok {
				multi_nodes_map[node.ID] = node
				multi_nodes = append(multi_nodes, node)
			}

			hostids = append(hostids, node.ID)
		}
	}

	for _, edge := range edges.Edges {
		if edge.Type == "tcp" || edge.Type == "udp" {
			if _, ok := multi_edges_map[edge.ID]; !ok {
				multi_edges_map[edge.ID] = edge
				multi_edges = append(multi_edges, edge)
			}

			if _, ok := multi_nodes_map[nodes.Lookup[edge.Src].ID]; !ok {
				multi_nodes_map[nodes.Lookup[edge.Src].ID] = nodes.Lookup[edge.Src]
				multi_nodes = append(multi_nodes, nodes.Lookup[edge.Src])
			}

			if _, ok := multi_nodes_map[nodes.Lookup[edge.Dst].ID]; !ok {
				multi_nodes_map[nodes.Lookup[edge.Dst].ID] = nodes.Lookup[edge.Dst]
				multi_nodes = append(multi_nodes, nodes.Lookup[edge.Dst])
			}
		} else if edge.Type == "server" || edge.Type == "client" {
			if _, ok := multi_edges_map[edge.ID]; !ok {
				multi_edges_map[edge.ID] = edge
				multi_edges = append(multi_edges, edge)
			}

			if _, ok := multi_nodes_map[nodes.Lookup[edge.Src].ID]; !ok {
				multi_nodes_map[nodes.Lookup[edge.Src].ID] = nodes.Lookup[edge.Src]
				multi_nodes = append(multi_nodes, nodes.Lookup[edge.Src])
			}

			if _, ok := multi_nodes_map[nodes.Lookup[edge.Dst].ID]; !ok {
				multi_nodes_map[nodes.Lookup[edge.Dst].ID] = nodes.Lookup[edge.Dst]
				multi_nodes = append(multi_nodes, nodes.Lookup[edge.Dst])
			}

			// 创建 net 节点相连的 process 节点与 host 节点的边实例
			for _, hostid := range hostids {
				if nodes.Lookup[edge.Dst].UUID == nodes.Lookup[hostid].UUID {
					net_process__host_edge := &meta.Edge{
						ID:   fmt.Sprintf("%s_%s_%s", edge.Dst, meta.EDGE_BELONG, hostid),
						Type: meta.EDGE_BELONG,
						Src:  edge.Dst,
						Dst:  hostid,
						Dir:  true,
					}

					if _, ok := multi_edges_map[net_process__host_edge.ID]; !ok {
						multi_edges_map[net_process__host_edge.ID] = net_process__host_edge
						multi_edges = append(multi_edges, net_process__host_edge)
					}

					break
				}
			}
		}
	}
	return multi_nodes, multi_edges, collect_errlist, process_errlist
}
