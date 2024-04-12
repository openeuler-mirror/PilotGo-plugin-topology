package generator

import (
	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/graph"
)

type TopoInterface interface {
	CreateNodeEntities(*agentmanager.Agent, *graph.Nodes) error
	CreateEdgeEntities(*agentmanager.Agent, *graph.Edges, *graph.Nodes) error
	Return_Agent_node_count() *int32
}
