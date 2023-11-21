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
		Unixtime: neo4jnode.Props["unixtime"].(string),
		Tags:     neo4jnode.Labels,
		Metrics:  metrics,
	}

	switch neo4jnode.Labels[0] {
	case meta.NODE_APP:
		toponode.Type = neo4jnode.Labels[0]
		toponode.UUID = neo4jnode.Labels[1]
	case meta.NODE_HOST:
		toponode.Type = neo4jnode.Labels[0]
		toponode.UUID = neo4jnode.Labels[1]
	case meta.NODE_NET:
		toponode.Type = neo4jnode.Labels[0]
		toponode.UUID = neo4jnode.Labels[1]
	case meta.NODE_PROCESS:
		toponode.Type = neo4jnode.Labels[0]
		toponode.UUID = neo4jnode.Labels[1]
	case meta.NODE_RESOURCE:
		toponode.Type = neo4jnode.Labels[0]
		toponode.UUID = neo4jnode.Labels[1]
	case meta.NODE_THREAD:
		toponode.Type = neo4jnode.Labels[0]
		toponode.UUID = neo4jnode.Labels[1]
	default:
		toponode.Type = neo4jnode.Labels[1]
		toponode.UUID = neo4jnode.Labels[0]
	}

	return toponode
}

func Neo4jrelaToToporela(neo4jrela neo4j.Relationship) *meta.Edge {
	metrics := make(map[string]string)

	for k, v := range neo4jrela.Props {
		metrics[k] = v.(string)
	}

	tags := []string{}
	tags = append(tags, neo4jrela.Type)

	toporela := &meta.Edge{
		DBID:     neo4jrela.Id,
		SrcID:    neo4jrela.StartId,
		DstID:    neo4jrela.EndId,
		Src:      neo4jrela.Props["src"].(string),
		Dst:      neo4jrela.Props["dst"].(string),
		ID:       neo4jrela.Props["rid"].(string),
		Type:     neo4jrela.Type,
		Dir:      neo4jrela.Props["dir"].(string),
		Unixtime: neo4jrela.Props["unixtime"].(string),
		Tags:     tags,
		Metrics:  metrics,
	}

	return toporela
}
