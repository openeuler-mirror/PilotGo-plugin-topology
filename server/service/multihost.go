package service

import (
	"fmt"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"github.com/pkg/errors"
)

func MultiHostService() ([]*meta.Node, []*meta.Edge, []map[string]string, error) {
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
	combos := make([]map[string]string, 0)

	if dao.Global_GraphDB == nil {
		err := errors.New("dao.global_graphdb is nil **warn**1")
		return nil, nil, nil, err
	}

	times, err := dao.Global_GraphDB.Timestamps_query()
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, nil, nil, err
	}

	if len(times) != 0 {
		if len(times) < 2 {
			latest = times[0]
		} else {
			latest = times[len(times)-2]
		}
	} else {
		err := errors.New("the number of timestamp is zero **warn**0")
		return nil, nil, nil, err
	}

	nodes, err = dao.Global_GraphDB.MultiHost_node_query(latest)
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, nil, nil, err
	}

	for _, _node := range nodes {
		nodes_map[_node.DBID] = _node
	}

	edges, err = dao.Global_GraphDB.MultiHost_relation_query(latest)
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, nil, nil, err
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

			combos = append(combos, map[string]string{
				"id":    node.UUID,
				"label": node.UUID,
			})
		}
	}

	for _, edge := range edges {
		if edge.Type == "tcp" || edge.Type == "udp" {
			if _, ok := multi_edges_map[edge.DBID]; !ok {
				multi_edges_map[edge.DBID] = edge
				multi_edges = append(multi_edges, edge)
			}

			if _, ok1 := multi_nodes_map[edge.SrcID]; !ok1 {
				if n, ok2 := nodes_map[edge.SrcID]; ok2 {
					multi_nodes_map[edge.SrcID] = n
					multi_nodes = append(multi_nodes, n)
				}
			}

			if _, ok1 := multi_nodes_map[edge.DstID]; !ok1 {
				if n, ok2 := nodes_map[edge.DstID]; ok2 {
					multi_nodes_map[edge.DstID] = n
					multi_nodes = append(multi_nodes, n)
				}
			}
		} else if edge.Type == "server" || edge.Type == "client" {
			if _, ok := multi_edges_map[edge.DBID]; !ok {
				multi_edges_map[edge.DBID] = edge
				multi_edges = append(multi_edges, edge)
			}

			if _, ok1 := multi_nodes_map[edge.SrcID]; !ok1 {
				if n, ok2 := nodes_map[edge.SrcID]; ok2 {
					multi_nodes_map[edge.SrcID] = n
					multi_nodes = append(multi_nodes, n)
				}
			}

			if _, ok1 := multi_nodes_map[edge.DstID]; !ok1 {
				if n, ok2 := nodes_map[edge.DstID]; ok2 {
					multi_nodes_map[edge.DstID] = n
					multi_nodes = append(multi_nodes, n)
				}
			}

			// 创建 net 节点相连的 process 节点与 host 节点的边实例
			for _, hostid := range hostids {
				process_node, ok1 := nodes_map[edge.DstID]
				host_node, ok2 := nodes_map[hostid]
				if ok1 && ok2 && process_node.UUID == host_node.UUID {
					net_process_host_edge := &meta.Edge{
						ID:   fmt.Sprintf("%s_%s_%s", process_node.ID, meta.EDGE_BELONG, host_node.ID),
						Type: meta.EDGE_BELONG,
						Src:  edge.Dst,
						Dst:  host_node.ID,
						Dir:  "direct",
					}

					net_process_host_edge.Tags = append(net_process_host_edge.Tags, meta.EDGE_BELONG)

					// TODO: multi_edges_map未包含新创建的边, multi_edges中新创建的边没有DBID、SrcID、DstID
					// 针对机器中某个process节点存在多个net节点的情况，在创建process-host边时去掉重复的边
					repeat := false
					for _, edge_in_multi := range multi_edges {
						if net_process_host_edge.ID == edge_in_multi.ID {
							repeat = true
							break
						}
					}
					if !repeat {
						multi_edges = append(multi_edges, net_process_host_edge)
					}

					break
				}
			}
		}
	}

	return multi_nodes, multi_edges, combos, nil
}
