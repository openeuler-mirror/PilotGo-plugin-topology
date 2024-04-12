package service

import (
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/graphmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/graph"
	"github.com/pkg/errors"
)

func SingleHostTreeService(uuid string) (*graph.TreeTopoNode, error) {
	var latest string
	var treerootnode *graph.TreeTopoNode
	var single_nodes []*graph.Node
	single_nodes_map := make(map[int64]*graph.Node)
	treenodes_process := make([]*graph.TreeTopoNode, 0)
	treenodes_net := make([]*graph.TreeTopoNode, 0)
	nodes_type_map := make(map[string][]*graph.Node)

	if graphmanager.Global_GraphDB == nil {
		err := errors.New("global_graphdb is nil **errstackfatal**1")
		return nil, err
	}

	times, err := graphmanager.Global_GraphDB.Timestamps_query()
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, err
	}

	if len(times) != 0 {
		if len(times) < 2 {
			latest = times[0]
		} else {
			latest = times[len(times)-2]
		}
	} else {
		err := errors.New("the number of timestamp is zero **errstack**0")
		return nil, err
	}

	single_nodes, err = graphmanager.Global_GraphDB.SingleHost_node_query(uuid, latest)
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, err
	}

	for _, node := range single_nodes {
		if _, ok := single_nodes_map[node.DBID]; !ok {
			single_nodes_map[node.DBID] = node
		}
	}

	for _, node := range single_nodes {
		nodes_type_map[node.Type] = append(nodes_type_map[node.Type], node)
		if node.Type == "host" {
			treerootnode = graph.CreateTreeNode(node)
		}
	}

	if treerootnode == nil {
		err := errors.New("there are no host node in single_nodes **errstack**5")
		return nil, err
	}

	for _, node := range nodes_type_map[global.NODE_RESOURCE] {
		childnode := graph.CreateTreeNode(node)
		treerootnode.Children = append(treerootnode.Children, childnode)
	}

	for _, node := range nodes_type_map[global.NODE_PROCESS] {
		treenode := graph.CreateTreeNode(node)
		treenodes_process = append(treenodes_process, treenode)
	}

	for _, node := range nodes_type_map[global.NODE_NET] {
		treenode := graph.CreateTreeNode(node)
		treenodes_net = append(treenodes_net, treenode)
	}

	for _, node := range treenodes_process {
		if node.Node.Metrics["Pid"] == "1" {
			node.Children = graph.SliceToTree(treenodes_process, treenodes_net, "1")
			treerootnode.Children = append(treerootnode.Children, node)

			break
		}
	}

	return treerootnode, nil
}
