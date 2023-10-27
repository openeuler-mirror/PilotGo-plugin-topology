package utils

import (
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// neo4jnode to toponode
func Neo4jnodeToToponode(neo4jnode neo4j.Node) *meta.Node {
	metrics := make(map[string]string)

	for k, v := range neo4jnode.Props {
		metrics[k] = v.(string)
	}

	toponode := &meta.Node{
		DBID:     neo4jnode.Id,
		ID:       neo4jnode.Props["nid"].(string),
		Name:     neo4jnode.Props["name"].(string),
		Type:     neo4jnode.Labels[0],
		UUID:     neo4jnode.Labels[1],
		Unixtime: neo4jnode.Props["unixtime"].(string),
		Metrics:  metrics,
	}

	return toponode
}

func Neo4jrelaToToporela(neo4jrela neo4j.Relationship) *meta.Edge {
	metrics := make(map[string]string)

	for k, v := range neo4jrela.Props {
		metrics[k] = v.(string)
	}

	toporela := &meta.Edge{
		DBID:     neo4jrela.Id,
		Src:      neo4jrela.StartId,
		Dst:      neo4jrela.EndId,
		ID:       neo4jrela.Props["rid"].(string),
		Type:     neo4jrela.Type,
		Dir:      neo4jrela.Props["dir"].(bool),
		Unixtime: neo4jrela.Props["unixtime"].(string),
		Metrics:  metrics,
	}

	return toporela
}
