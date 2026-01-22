package client

import (
	"encoding/json"

	"github.com/skydive-project/skydive/flow"
	gclient "github.com/skydive-project/skydive/graffiti/api/client"
	"github.com/skydive-project/skydive/sflow"
	"github.com/skydive-project/skydive/topology"
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

// GetInterfaceMetrics from Gremlin query
func (g *GremlinQueryHelper) GetInterfaceMetrics(query interface{}) (map[string][]*topology.InterfaceMetric, error) {
	data, err := g.Query(query)
	if err != nil {
		return nil, err
	}

	var result []map[string][]*topology.InterfaceMetric
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result[0], nil
}

// GetSFlowMetrics from Gremlin query
func (g *GremlinQueryHelper) GetSFlowMetrics(query interface{}) (map[string][]*sflow.SFMetric, error) {
	data, err := g.Query(query)
	if err != nil {
		return nil, err
	}

	var result []map[string][]*sflow.SFMetric
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result[0], nil
}

// GetFlowMetrics from Gremlin query
func (g *GremlinQueryHelper) GetFlowMetrics(query interface{}) (map[string][]*flow.FlowMetric, error) {
	data, err := g.Query(query)
	if err != nil {
		return nil, err
	}

	var result []map[string][]*flow.FlowMetric
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrNoResult
	}

	return result[0], nil
}

// GetFlowMetric from Gremlin query
func (g *GremlinQueryHelper) GetFlowMetric(query interface{}) (*flow.FlowMetric, error) {
	data, err := g.Query(query)
	if err != nil {
		return nil, err
	}

	var result flow.FlowMetric
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
