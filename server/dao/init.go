package dao

import "gitee.com/openeuler/PilotGo-plugin-topology-server/meta"

var Global_GraphDB GraphdbIface

type GraphdbIface interface {
	ClearExpiredData(int64)
	
	Node_create(string, *meta.Node) error
	Edge_create(string, *meta.Edge) error

	Timestamps_query() ([]string, error)

	Node_query(string, string) ([]*meta.Node, error)
	SingleHost_node_query(string, string) ([]*meta.Node, error)
	MultiHost_node_query(string) ([]*meta.Node, error)

	Relation_query(string, string) ([]*meta.Edge, error)
	MultiHost_relation_query(string) ([]*meta.Edge, error)
}
