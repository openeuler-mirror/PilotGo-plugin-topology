/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package public

import (
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/db/graphmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/graph"
	"github.com/pkg/errors"
)

func SingleHostTreeService(uuid string) (*graph.TreeTopoNode, error, bool) {
	var latest string
	var treerootnode *graph.TreeTopoNode
	var single_nodes []*graph.Node
	single_nodes_map := make(map[int64]*graph.Node)
	treenodes_process := make([]*graph.TreeTopoNode, 0)
	treenodes_net := make([]*graph.TreeTopoNode, 0)
	nodes_type_map := make(map[string][]*graph.Node)

	if graphmanager.Global_GraphDB == nil {
		err := errors.New("global_graphdb is nil")
		return nil, err, true
	}

	times, err := graphmanager.Global_GraphDB.Timestamps_query()
	if err != nil {
		err = errors.Wrap(err, " ")
		return nil, err, false
	}

	if len(times) != 0 {
		if len(times) < 2 {
			latest = times[0]
		} else {
			latest = times[len(times)-2]
		}
	} else {
		err := errors.New("the number of timestamp is zero")
		return nil, err, false
	}

	single_nodes, err = graphmanager.Global_GraphDB.SingleHost_node_query(uuid, latest)
	if err != nil {
		err = errors.Wrap(err, " ")
		return nil, err, false
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
		err := errors.New("there are no host node in single_nodes")
		return nil, err, false
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

	return treerootnode, nil, false
}
