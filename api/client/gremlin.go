package client

import (
	"encoding/json"

	"github.com/skydive-project/skydive/flow"
	gclient "github.com/skydive-project/skydive/graffiti/api/client"
)

// ErrNoResult is returned when a query returned no result
var ErrNoResult = gclient.ErrNoResult

// GremlinQueryHelper describes a gremlin query request query helper mechanism
type GremlinQueryHelper struct {
	*gclient.GremlinQueryHelper
}

// GetFlows from the Gremlin query
func (g *GremlinQueryHelper) GetFlows(query interface{}) ([]*flow.Flow, error) {
	data, err := g.Query(query)
	if err != nil {
		return nil, err
	}

	var flows []*flow.Flow
	if err := json.Unmarshal(data, &flows); err != nil {
		return nil, err
	}

	return flows, nil
}
