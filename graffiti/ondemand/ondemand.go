package ondemand

import (
	json "encoding/json"

	"github.com/skydive-project/skydive/graffiti/graph"
)

// Namespace "OnDemand"
const Namespace = "OnDemand"

// Query describes an ondemand query
// easyjson:json
type Query struct {
	NodeID   graph.Identifier
	Resource interface{}
}

// RawQuery describes a raw ondemand query
// easyjson:json
type RawQuery struct {
	NodeID   graph.Identifier
	Resource json.RawMessage
}

// Task describes the processing unit created by an ondemand
type Task = interface{}
