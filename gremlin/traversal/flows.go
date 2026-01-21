package traversal

import (
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
