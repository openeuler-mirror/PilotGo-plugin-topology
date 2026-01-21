package traversal

import (
	"errors"
	"fmt"

	"atomgit.com/openeuler/PilotGo-plugin-topology/flow"
	"atomgit.com/openeuler/PilotGo-plugin-topology/flow/storage"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/filters"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph/traversal"
)

const (
	defaultSortBy = "Last"
)

// FlowTraversalExtension describes flows in a graph Gremlin language extension
type FlowTraversalExtension struct {
	FlowToken        traversal.Token
	HopsToken        traversal.Token
	NodesToken       traversal.Token
	CaptureNodeToken traversal.Token
	AggregatesToken  traversal.Token
	GroupToken       traversal.Token
	BpfToken         traversal.Token
	TableClient      flow.TableClient
	Storage          storage.Storage
}

// FlowGremlinTraversalStep a flow Gremlin language step
type FlowGremlinTraversalStep struct {
	TableClient        flow.TableClient
	Storage            storage.Storage
	context            traversal.GremlinTraversalContext
	paramsFilter       *filters.Filter
	metricsNextStep    bool
	rawpacketsNextStep bool
	dedup              bool
	dedupBy            string
	sort               bool
	sortBy             string
	sortOrder          filters.SortOrder
}

// FlowTraversalStep a flow step linked to a storage
type FlowTraversalStep struct {
	GraphTraversal  *traversal.GraphTraversal
	Storage         storage.Storage
	flowset         *flow.FlowSet
	flowSearchQuery filters.SearchQuery
	error           error
}

// HopsGremlinTraversalStep hops step
type HopsGremlinTraversalStep struct {
	traversal.GremlinTraversalContext
}

// NodesGremlinTraversalStep nodes step
type NodesGremlinTraversalStep struct {
	traversal.GremlinTraversalContext
}

// CaptureNodeGremlinTraversalStep capture step
type CaptureNodeGremlinTraversalStep struct {
	traversal.GremlinTraversalContext
}

// AggregatesGremlinTraversalStep aggregates step
type AggregatesGremlinTraversalStep struct {
	traversal.GremlinTraversalContext
}

// BpfGremlinTraversalStep step
type BpfGremlinTraversalStep struct {
	traversal.GremlinTraversalContext
}

// Out returns the B node
func (f *FlowTraversalStep) Out(ctx traversal.StepContext, s ...interface{}) *traversal.GraphTraversalV {
	var nodes []*graph.Node

	if f.error != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, f.error)
	}

	filter1, err := traversal.ParamsToFilter(filters.BoolFilterOp_AND, s...)
	if err != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
	}

	it := ctx.PaginationRange.Iterator()

	f.GraphTraversal.RLock()
	defer f.GraphTraversal.RUnlock()

	for _, flow := range f.flowset.Flows {
		if it.Done() {
			break
		}
		if flow.Link.B != "" {
			f1, err := traversal.KeyValueToFilter("MAC", flow.Link.B)
			if err != nil {
				return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
			}
			f2, err := traversal.KeyValueToFilter("PeerIntfMAC", flow.Link.B)
			if err != nil {
				return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
			}
			filter2 := filters.NewOrFilter(f1, f2)
			matcher := graph.NewElementFilter(filters.NewAndFilter(filter1, filter2))

			if node := f.GraphTraversal.Graph.LookupFirstNode(matcher); node != nil && it.Next() {
				nodes = append(nodes, node)
			}
		}
	}

	return traversal.NewGraphTraversalV(f.GraphTraversal, nodes)
}

// In returns the A node
func (f *FlowTraversalStep) In(ctx traversal.StepContext, s ...interface{}) *traversal.GraphTraversalV {
	var nodes []*graph.Node

	if f.error != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, f.error)
	}

	filter1, err := traversal.ParamsToFilter(filters.BoolFilterOp_AND, s...)
	if err != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
	}

	it := ctx.PaginationRange.Iterator()

	f.GraphTraversal.RLock()
	defer f.GraphTraversal.RUnlock()

	for _, flow := range f.flowset.Flows {
		if it.Done() {
			break
		}

		if flow.Link.A != "" {
			f1, err := traversal.KeyValueToFilter("MAC", flow.Link.A)
			if err != nil {
				return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
			}
			f2, err := traversal.KeyValueToFilter("PeerIntfMAC", flow.Link.A)
			if err != nil {
				return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
			}
			filter2 := filters.NewOrFilter(f1, f2)
			matcher := graph.NewElementFilter(filters.NewAndFilter(filter1, filter2))

			if node := f.GraphTraversal.Graph.LookupFirstNode(matcher); node != nil && it.Next() {
				nodes = append(nodes, node)
			}
		}
	}

	return traversal.NewGraphTraversalV(f.GraphTraversal, nodes)
}

// Both returns A and B nodes
func (f *FlowTraversalStep) Both(ctx traversal.StepContext, s ...interface{}) *traversal.GraphTraversalV {
	var nodes []*graph.Node

	if f.error != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, f.error)
	}

	filter1, err := traversal.ParamsToFilter(filters.BoolFilterOp_AND, s...)
	if err != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
	}

	it := ctx.PaginationRange.Iterator()

	f.GraphTraversal.RLock()
	defer f.GraphTraversal.RUnlock()

	for _, flow := range f.flowset.Flows {
		if it.Done() {
			break
		}

		if flow.Link.A != "" {
			f1, err := traversal.KeyValueToFilter("MAC", flow.Link.A)
			if err != nil {
				return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
			}
			f2, err := traversal.KeyValueToFilter("PeerIntfMAC", flow.Link.A)
			if err != nil {
				return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
			}
			filter2 := filters.NewOrFilter(f1, f2)

			matcher := graph.NewElementFilter(filters.NewAndFilter(filter1, filter2))

			if node := f.GraphTraversal.Graph.LookupFirstNode(matcher); node != nil && it.Next() {
				nodes = append(nodes, node)
			}
		}
		if flow.Link.B != "" {
			f1, err := traversal.KeyValueToFilter("MAC", flow.Link.B)
			if err != nil {
				return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
			}
			f2, err := traversal.KeyValueToFilter("PeerIntfMAC", flow.Link.B)
			if err != nil {
				return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
			}
			filter2 := filters.NewOrFilter(f1, f2)
			matcher := graph.NewElementFilter(filters.NewAndFilter(filter1, filter2))

			if node := f.GraphTraversal.Graph.LookupFirstNode(matcher); node != nil && it.Next() {
				nodes = append(nodes, node)
			}
		}
	}

	return traversal.NewGraphTraversalV(f.GraphTraversal, nodes)
}

// Nodes returns A, B and the capture nodes
func (f *FlowTraversalStep) Nodes(ctx traversal.StepContext, s ...interface{}) *traversal.GraphTraversalV {
	var nodes []*graph.Node

	if f.error != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, f.error)
	}

	filter1, err := traversal.ParamsToFilter(filters.BoolFilterOp_AND, s...)
	if err != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
	}

	it := ctx.PaginationRange.Iterator()

	f.GraphTraversal.RLock()
	defer f.GraphTraversal.RUnlock()

	for _, flow := range f.flowset.Flows {
		if it.Done() {
			break
		}

		if flow.NodeTID != "" {
			filter := filters.NewAndFilter(filter1, filters.NewTermStringFilter("TID", flow.NodeTID))
			if node := f.GraphTraversal.Graph.LookupFirstNode(graph.NewElementFilter(filter)); node != nil && it.Next() {
				nodes = append(nodes, node)
			}
		}

		if flow.Link.A != "" {
			filter := filters.NewAndFilter(filter1, filters.NewTermStringFilter("MAC", flow.Link.A))
			if node := f.GraphTraversal.Graph.LookupFirstNode(graph.NewElementFilter(filter)); node != nil && it.Next() {
				nodes = append(nodes, node)
			}
		}
		if flow.Link.B != "" {
			filter := filters.NewAndFilter(filter1, filters.NewTermStringFilter("MAC", flow.Link.B))
			if node := f.GraphTraversal.Graph.LookupFirstNode(graph.NewElementFilter(filter)); node != nil && it.Next() {
				nodes = append(nodes, node)
			}
		}
	}
	return traversal.NewGraphTraversalV(f.GraphTraversal, nodes)
}

// Hops returns all the capture nodes where the flow was seen
func (f *FlowTraversalStep) Hops(ctx traversal.StepContext, s ...interface{}) *traversal.GraphTraversalV {
	var nodes []*graph.Node

	if f.error != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, f.error)
	}

	filter1, err := traversal.ParamsToFilter(filters.BoolFilterOp_AND, s...)
	if err != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
	}

	it := ctx.PaginationRange.Iterator()

	f.GraphTraversal.RLock()
	defer f.GraphTraversal.RUnlock()

	for _, fl := range f.flowset.Flows {
		if it.Done() {
			break
		}

		filter := filters.NewAndFilter(filter1, filters.NewTermStringFilter("TID", fl.NodeTID))
		if node := f.GraphTraversal.Graph.LookupFirstNode(graph.NewElementFilter(filter)); node != nil && it.Next() {
			nodes = append(nodes, node)
		}
	}

	return traversal.NewGraphTraversalV(f.GraphTraversal, nodes)
}

// Count step
func (f *FlowTraversalStep) Count(ctx traversal.StepContext, s ...interface{}) *traversal.GraphTraversalValue {
	if f.error != nil {
		return traversal.NewGraphTraversalValueFromError(f.error)
	}

	return traversal.NewGraphTraversalValue(f.GraphTraversal, len(f.flowset.Flows))
}

func paramsToFilter(filterOp filters.BoolFilterOp, params ...interface{}) (*filters.Filter, error) {
	if len(params) < 2 {
		return nil, errors.New("At least two parameters must be provided")
	}

	if len(params)%2 != 0 {
		return nil, fmt.Errorf("slice must be defined by pair k,v: %v", params)
	}

	var allFilters []*filters.Filter
	for i := 0; i < len(params); i += 2 {
		var filter *filters.Filter

		k, ok := params[i].(string)
		if !ok {
			return nil, errors.New("keys should be of string type")
		}

		if k == "Network" || k == "Link" || k == "Transport" {
			fa, err := traversal.KeyValueToFilter(k+".A", params[i+1])
			if err != nil {
				return nil, err
			}

			fb, err := traversal.KeyValueToFilter(k+".B", params[i+1])
			if err != nil {
				return nil, err
			}

			filter = filters.NewOrFilter(fa, fb)
		} else {
			f, err := traversal.KeyValueToFilter(k, params[i+1])
			if err != nil {
				return nil, err
			}
			filter = f
		}

		allFilters = append(allFilters, filter)
	}

	return filters.NewBoolFilter(filterOp, allFilters...), nil
}

func (f *FlowTraversalStep) has(filterOp filters.BoolFilterOp, ctx traversal.StepContext, s ...interface{}) *FlowTraversalStep {
	if f.error != nil {
		return f
	}

	filter, err := paramsToFilter(filterOp, s...)
	if err != nil {
		return &FlowTraversalStep{error: err}
	}

	return &FlowTraversalStep{GraphTraversal: f.GraphTraversal, Storage: f.Storage, flowset: f.flowset.Filter(filter)}
}

// Has step
func (f *FlowTraversalStep) Has(ctx traversal.StepContext, s ...interface{}) *FlowTraversalStep {
	return f.has(filters.BoolFilterOp_AND, ctx, s...)
}

// HasEither step
func (f *FlowTraversalStep) HasEither(ctx traversal.StepContext, s ...interface{}) *FlowTraversalStep {
	return f.has(filters.BoolFilterOp_OR, ctx, s...)
}

// Dedup deduplicate step
func (f *FlowTraversalStep) Dedup(ctx traversal.StepContext, keys ...interface{}) *FlowTraversalStep {
	if f.error != nil {
		return f
	}

	var key string
	if len(keys) > 0 {
		k, ok := keys[0].(string)
		if !ok {
			return &FlowTraversalStep{error: fmt.Errorf("Dedup parameter has to be a string key")}
		}
		key = k
	}

	if err := f.flowset.Dedup(key); err != nil {
		return &FlowTraversalStep{error: err}
	}

	return &FlowTraversalStep{GraphTraversal: f.GraphTraversal, Storage: f.Storage, flowset: f.flowset}
}

// CaptureNode step
func (f *FlowTraversalStep) CaptureNode(ctx traversal.StepContext, s ...interface{}) *traversal.GraphTraversalV {
	var nodes []*graph.Node

	if f.error != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, f.error)
	}

	filter1, err := traversal.ParamsToFilter(filters.BoolFilterOp_AND, s...)
	if err != nil {
		return traversal.NewGraphTraversalV(f.GraphTraversal, nodes, err)
	}

	it := ctx.PaginationRange.Iterator()

	f.GraphTraversal.RLock()
	defer f.GraphTraversal.RUnlock()

	for _, fl := range f.flowset.Flows {
		if it.Done() {
			break
		}

		filter := filters.NewAndFilter(filter1, filters.NewTermStringFilter("TID", fl.NodeTID))
		if node := f.GraphTraversal.Graph.LookupFirstNode(graph.NewElementFilter(filter)); node != nil && it.Next() {
			nodes = append(nodes, node)
		}
	}
	return traversal.NewGraphTraversalV(f.GraphTraversal, nodes)
}
