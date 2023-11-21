package service

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"github.com/pkg/errors"
)

func MultiHostService() ([]*meta.Node, []*meta.Edge, error) {
	var latest string
	nodes := make([]*meta.Node, 0)
	nodes_map := make(map[int64]*meta.Node)
	edges := make([]*meta.Edge, 0)
	edges_map := make(map[int64]*meta.Edge)
	hostids := make([]int64, 0)
	multi_nodes_map := make(map[int64]*meta.Node)
	multi_nodes := make([]*meta.Node, 0)
	multi_edges_map := make(map[int64]*meta.Edge)
	multi_edges := make([]*meta.Edge, 0)

	times, err := dao.Global_GraphDB.Timestamps_query()
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, nil, err
	}

	if len(times) < 2 {
		latest = times[0]
	} else {
		latest = times[len(times)-2]
	}

	nodes, err = dao.Global_GraphDB.MultiHost_node_query(latest)
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, nil, err
	}

	for _, _node := range nodes {
		nodes_map[_node.DBID] = _node
	}

	edges, err = dao.Global_GraphDB.MultiHost_relation_query(latest)
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, nil, err
	}

	for _, _edge := range edges {
		edges_map[_edge.DBID] = _edge
	}

	// 添加 host node
	for _, node := range nodes {
		if node.Type == "host" {
			if _, ok := multi_nodes_map[node.DBID]; !ok {
				multi_nodes_map[node.DBID] = node
				multi_nodes = append(multi_nodes, node)
			}

			hostids = append(hostids, node.DBID)
		}
	}

	for _, edge := range edges {
		if edge.Type == "tcp" || edge.Type == "udp" {
			if _, ok := multi_edges_map[edge.DBID]; !ok {
				multi_edges_map[edge.DBID] = edge
				multi_edges = append(multi_edges, edge)
			}

			if _, ok := multi_nodes_map[edge.SrcID]; !ok {
				multi_nodes_map[edge.SrcID] = nodes_map[edge.SrcID]
				multi_nodes = append(multi_nodes, nodes_map[edge.SrcID])
			}

			if _, ok := multi_nodes_map[edge.DstID]; !ok {
				multi_nodes_map[edge.DstID] = nodes_map[edge.DstID]
				multi_nodes = append(multi_nodes, nodes_map[edge.DstID])
			}
		} else if edge.Type == "server" || edge.Type == "client" {
			if _, ok := multi_edges_map[edge.DBID]; !ok {
				multi_edges_map[edge.DBID] = edge
				multi_edges = append(multi_edges, edge)
			}

			if _, ok := multi_nodes_map[edge.SrcID]; !ok {
				multi_nodes_map[edge.SrcID] = nodes_map[edge.SrcID]
				multi_nodes = append(multi_nodes, nodes_map[edge.SrcID])
			}

			if _, ok := multi_nodes_map[edge.DstID]; !ok {
				multi_nodes_map[edge.DstID] = nodes_map[edge.DstID]
				multi_nodes = append(multi_nodes, nodes_map[edge.DstID])
			}

			// 创建 net 节点相连的 process 节点与 host 节点的边实例
			for _, hostid := range hostids {
				if nodes_map[edge.DstID].UUID == nodes_map[hostid].UUID {
					net_process_host_edge := &meta.Edge{
						ID:   fmt.Sprintf("%s_%s_%s", nodes_map[edge.DstID].ID, meta.EDGE_BELONG, nodes_map[hostid].ID),
						Type: meta.EDGE_BELONG,
						Src:  edge.Dst,
						Dst:  nodes_map[hostid].ID,
						Dir:  "direct",
					}

					net_process_host_edge.Tags = append(net_process_host_edge.Tags, meta.EDGE_BELONG)
					
					// TODO: multi_edges_map未包含新创建的边, multi_edges中新创建的边没有DBID、SrcID、DstID
					// multi_edges_map[net_process__host_edge.ID] = net_process__host_edge
					multi_edges = append(multi_edges, net_process_host_edge)

					break
				}
			}
		}
	}

	return multi_nodes, multi_edges, nil
}
