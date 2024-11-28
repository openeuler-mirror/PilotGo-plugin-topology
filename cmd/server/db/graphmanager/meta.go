/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package graphmanager

import "gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/graph"

var Global_GraphDB GraphdbIface

type GraphdbIface interface {
	ClearExpiredData(int64)
	
	Node_create(string, *graph.Node) error
	Edge_create(string, *graph.Edge) error

	Timestamps_query() ([]string, error)

	Node_query(string, string) ([]*graph.Node, error)
	SingleHost_node_query(string, string) ([]*graph.Node, error)
	MultiHost_node_query(string) ([]*graph.Node, error)

	Relation_query(string, string) ([]*graph.Edge, error)
	MultiHost_relation_query(string) ([]*graph.Edge, error)
}
