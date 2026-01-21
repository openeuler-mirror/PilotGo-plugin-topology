package traversal

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"atomgit.com/openeuler/PilotGo-plugin-topology/common"
	"atomgit.com/openeuler/PilotGo-plugin-topology/flow"
	"atomgit.com/openeuler/PilotGo-plugin-topology/flow/storage"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/filters"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/getter"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph/traversal"
	"atomgit.com/openeuler/PilotGo-plugin-topology/topology/probes/socketinfo"
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

// Sort step
func (f *FlowTraversalStep) Sort(ctx traversal.StepContext, keys ...interface{}) *FlowTraversalStep {
	if f.error != nil {
		return f
	}

	sortOrder, sortBy, err := traversal.ParseSortParameter(keys...)
	if err != nil {
		return &FlowTraversalStep{error: err}
	}

	if sortBy == "" {
		sortBy = defaultSortBy
	}

	f.flowset.Sort(sortOrder, sortBy)
	return &FlowTraversalStep{GraphTraversal: f.GraphTraversal, Storage: f.Storage, flowset: f.flowset}
}

// Sum aggregates integer values mapped by 'key' cross flows
func (f *FlowTraversalStep) Sum(ctx traversal.StepContext, keys ...interface{}) *traversal.GraphTraversalValue {
	if f.error != nil {
		return traversal.NewGraphTraversalValueFromError(f.error)
	}

	if len(keys) != 1 {
		return traversal.NewGraphTraversalValueFromError(fmt.Errorf("Sum requires 1 parameter"))
	}

	key, ok := keys[0].(string)
	if !ok {
		return traversal.NewGraphTraversalValueFromError(fmt.Errorf("Sum parameter has to be a string key"))
	}

	k := strings.Split(key, ".")
	if k[0] != "Metric" && k[0] != "LastUpdateMetric" {
		return traversal.NewGraphTraversalValueFromError(fmt.Errorf("Sum accepts only sub fields of Metric and LastUpadteMetric"))
	}

	var s float64
	for _, fl := range f.flowset.Flows {
		if v, err := fl.GetFieldInt64(key); err == nil {
			s += float64(v)
		} else {
			return traversal.NewGraphTraversalValueFromError(err)
		}
	}
	return traversal.NewGraphTraversalValue(f.GraphTraversal, s)
}

// PropertyValues returns a flow field value
func (f *FlowTraversalStep) PropertyValues(ctx traversal.StepContext, keys ...interface{}) *traversal.GraphTraversalValue {
	if f.error != nil {
		return traversal.NewGraphTraversalValueFromError(f.error)
	}

	key := keys[0].(string)
	var s []interface{}
	for _, fl := range f.flowset.Flows {
		v, err := fl.GetField(key)
		if err != nil {
			return traversal.NewGraphTraversalValueFromError(getter.ErrFieldNotFound)
		}
		s = append(s, v)
	}
	return traversal.NewGraphTraversalValue(f.GraphTraversal, s)
}

// PropertyKeys returns flow fields
func (f *FlowTraversalStep) PropertyKeys(ctx traversal.StepContext, keys ...interface{}) *traversal.GraphTraversalValue {
	if f.error != nil {
		return traversal.NewGraphTraversalValueFromError(f.error)
	}

	var s []string

	if len(f.flowset.Flows) > 0 {
		// all Flow struct are the same, take the first one
		s = f.flowset.Flows[0].GetFieldKeys()
	}

	return traversal.NewGraphTraversalValue(f.GraphTraversal, s)
}

// FlowMetrics returns flow metric counters
func (f *FlowTraversalStep) FlowMetrics(ctx traversal.StepContext) *MetricsTraversalStep {
	if f.error != nil {
		return NewMetricsTraversalStepFromError(f.error)
	}

	var flowMetrics map[string][]common.Metric

	context := f.GraphTraversal.Graph.GetContext()
	if context.TimeSlice != nil {
		// two cases, either we have a flowset and we need to use it in order to filter
		// flows or we don't have flowset but we have the pre-built flowSearchQuery filter
		// if none of these cases it's an error.
		if f.flowset != nil {
			flowFilter := flow.NewFilterForFlowSet(f.flowset)
			f.flowSearchQuery.Filter = filters.NewAndFilter(f.flowSearchQuery.Filter, flowFilter)
		} else if f.flowSearchQuery.Filter == nil {
			return NewMetricsTraversalStepFromError(errors.New("Unable to filter flows"))
		}

		fr := filters.Range{To: context.TimeSlice.Last}
		if context.TimeSlice.Start != context.TimeSlice.Last {
			fr.From = context.TimeSlice.Start
		}
		metricFilter := filters.NewFilterIncludedIn(fr, "")

		f.flowSearchQuery.Sort = true
		f.flowSearchQuery.SortBy = defaultSortBy
		f.flowSearchQuery.SortOrder = filters.SortOrder_Ascending

		var err error
		if flowMetrics, err = f.Storage.SearchMetrics(f.flowSearchQuery, metricFilter); err != nil {
			return NewMetricsTraversalStepFromError(err)
		}
	} else {
		flowMetrics = make(map[string][]common.Metric, len(f.flowset.Flows))
		for _, f := range f.flowset.Flows {
			var metric *flow.FlowMetric
			if f.LastUpdateMetric != nil {
				metric = f.LastUpdateMetric
			} else {
				// if we get empty LastUpdateMetric it means that we got flow not already updated
				// by the flow table update ticker, so packets between the start of the flow and
				// the first update.
				metric = f.Metric
			}
			flowMetrics[f.UUID] = append(flowMetrics[f.UUID], metric)
		}
	}

	return NewMetricsTraversalStep(f.GraphTraversal, flowMetrics)
}

// RawPackets searches for RawPacket based on previous flow filter from
// either agents or datastore.
func (f *FlowTraversalStep) RawPackets(ctx traversal.StepContext) *RawPacketsTraversalStep {
	if f.error != nil {
		return &RawPacketsTraversalStep{error: f.error}
	}

	rawPackets := make(map[string][]*flow.RawPacket)

	context := f.GraphTraversal.Graph.GetContext()
	if context.TimeSlice != nil {
		// two cases, either we have a flowset and we need to use it in order to filter
		// flows or we don't have flowset but we have the pre-built flowSearchQuery filter
		// if none of these cases it's an error.
		if f.flowset != nil {
			flowFilter := flow.NewFilterForFlowSet(f.flowset)
			f.flowSearchQuery.Filter = filters.NewAndFilter(f.flowSearchQuery.Filter, flowFilter)
		} else if f.flowSearchQuery.Filter == nil {
			return &RawPacketsTraversalStep{error: errors.New("Unable to filter flows")}
		}

		fr := filters.Range{To: context.TimeSlice.Last}
		if context.TimeSlice.Start != context.TimeSlice.Last {
			fr.From = context.TimeSlice.Start
		}

		rawPacketsFilter := filters.NewAndFilter(
			filters.NewGteInt64Filter("Timestamp", fr.From),
			filters.NewLteInt64Filter("Timestamp", fr.To),
		)

		f.flowSearchQuery.Sort = true
		f.flowSearchQuery.SortBy = "Index"
		f.flowSearchQuery.SortOrder = filters.SortOrder_Ascending

		var err error
		if rawPackets, err = f.Storage.SearchRawPackets(f.flowSearchQuery, rawPacketsFilter); err != nil {
			return &RawPacketsTraversalStep{error: err}
		}
	} else {
		for _, fl := range f.flowset.Flows {
			if len(fl.LastRawPackets) > 0 {
				rawPackets[fl.UUID] = fl.LastRawPackets
			}
		}
	}

	return &RawPacketsTraversalStep{GraphTraversal: f.GraphTraversal, rawPackets: rawPackets}
}

// Sockets returns the sockets at both sides of the specified flows
func (f *FlowTraversalStep) Sockets(ctx traversal.StepContext, s ...interface{}) *SocketsTraversalStep {
	if f.error != nil {
		return &SocketsTraversalStep{error: f.error}
	}

	if len(s) != 0 {
		return &SocketsTraversalStep{error: fmt.Errorf("Sockets requires no parameter")}
	}

	indexer := NewSocketIndexer(f.GraphTraversal.Graph)

	flowSockets := make(map[string][]*socketinfo.ConnectionInfo)
	for _, fl := range f.flowset.Flows {
		if transport := fl.GetTransport(); transport != nil && (transport.GetProtocol() == flow.FlowProtocol_TCP || transport.GetProtocol() == flow.FlowProtocol_UDP) {
			var sockets []*socketinfo.ConnectionInfo

			localPort := fl.GetTransport().GetA()
			remotePort := fl.GetTransport().GetB()
			protocol := transport.GetProtocol()

			hash := socketinfo.HashTuple(protocol, net.ParseIP(fl.GetNetwork().GetA()), int64(localPort), net.ParseIP(fl.GetNetwork().GetB()), int64(remotePort))
			_, socks := indexer.FromHash(hash)
			for _, socket := range socks {
				sockets = append(sockets, socket.(*socketinfo.ConnectionInfo))
			}

			hash = socketinfo.HashTuple(protocol, net.ParseIP(fl.GetNetwork().GetB()), int64(remotePort), net.ParseIP(fl.GetNetwork().GetA()), int64(localPort))
			_, socks = indexer.FromHash(hash)
			for _, socket := range socks {
				sockets = append(sockets, socket.(*socketinfo.ConnectionInfo))
			}

			if len(sockets) > 0 {
				flowSockets[fl.UUID] = sockets
			}
		}
	}

	return &SocketsTraversalStep{GraphTraversal: f.GraphTraversal, sockets: flowSockets}
}
