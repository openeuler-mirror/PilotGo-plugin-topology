// Package types defines the API resource types
//
// swagger:meta
package types

import (
	"github.com/skydive-project/skydive/graffiti/api/rest"
	gtypes "github.com/skydive-project/skydive/graffiti/api/types"
	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/graffiti/schema"
)

// SchemaValidator validates resources against JSON schemas
var SchemaValidator schema.Validator

// Capture object
//
// Captures provide a way to capture network traffic on the nodes
// matching a Gremlin expression.
//
// easyjson:json
// swagger:model Capture
type Capture struct {
	// swagger:allOf
	rest.BasicResource `yaml:",inline"`
	// Gremlin Query
	// required: true
	GremlinQuery string `json:"GremlinQuery,omitempty" valid:"isGremlinExpr" yaml:"GremlinQuery"`
	// BPF filter
	BPFFilter string `json:"BPFFilter,omitempty" valid:"isBPFFilter" yaml:"BPFFilter"`
	// Capture name
	Name string `json:"Name,omitempty" yaml:"Name"`
	// Capture description
	Description string `json:"Description,omitempty" yaml:"Description"`
	// Capture type. Can be afpacket, pcap, ebpf, sflow, pcapsocket, ovsmirror, dpdk, ovssflow or ovsnetflow
	Type string `json:"Type,omitempty" valid:"isValidCaptureType" yaml:"Type"`
	// Number of active captures
	// swagger:ignore
	Count int `json:"Count" yaml:"Count"`
	// SFlow port
	Port int `json:"Port,omitempty" yaml:"Port"`
	// Sampling rate for SFlow flows. 0: no flow samples
	SamplingRate uint32 `json:"SamplingRate" yaml:"SamplingRate"`
	// Polling interval for SFlow counters, 0: no counter samples
	PollingInterval uint32 `json:"PollingInterval" yaml:"PollingInterval"`
	// Maximum number of raw packets captured, 0: no packet, -1: unlimited
	RawPacketLimit int `json:"RawPacketLimit,omitempty" valid:"isValidRawPacketLimit" yaml:"RawPacketLimit"`
	// Packet header size to consider
	HeaderSize int `json:"HeaderSize,omitempty" valid:"isValidCaptureHeaderSize" yaml:"HeaderSize"`
	// Add additional TCP metrics to flows
	ExtraTCPMetric bool `json:"ExtraTCPMetric" yaml:"ExtraTCPMetric"`
	// Defragment IPv4 packets
	IPDefrag bool `json:"IPDefrag" yaml:"IPDefrag"`
	// Reassemble TCP packets
	ReassembleTCP bool `json:"ReassembleTCP" yaml:"ReassembleTCP"`
	// First layer used by flow key calculation, L2 or L3
	LayerKeyMode string `json:"LayerKeyMode,omitempty" valid:"isValidLayerKeyMode" yaml:"LayerKeyMode"`
	// List of extra layers to be added to the flow, available: DNS|DHCPv4|VRRP
	ExtraLayers int `json:"ExtraLayers,omitempty" yaml:"ExtraLayers"`
	// sFlow/NetFlow target, if empty the agent will be used
	Target string `json:"Target,omitempty" valid:"isValidAddress" yaml:"Target"`
	// target type (netflowv5, erspanv1), ignored in case of sFlow/NetFlow capture
	TargetType string `json:"TargetType,omitempty" yaml:"TargetType"`
}

// GetName returns the resource name
func (c *Capture) GetName() string {
	return "Capture"
}

// NewCapture creates a new capture
func NewCapture(query string, bpfFilter string) *Capture {
	return &Capture{
		GremlinQuery: query,
		BPFFilter:    bpfFilter,
	}
}

// Edge object
// swagger:model
// easyjson:skip
type Edge = gtypes.Edge

// EdgeRule object
//
// Edge rules allow the dynamic creation of links between nodes of the graph.
//
// easyjson:json
// swagger:model
type EdgeRule struct {
	// swagger:allOf
	rest.BasicResource `yaml:",inline"`
	// Edge rule name
	Name string `yaml:"Name"`
	// Edge rule description
	Description string `yaml:"Description"`
	// Gremlin expression of the edges source nodes
	Src string `valid:"isGremlinExpr" yaml:"Src"`
	// Gremlin expression of the edges destination nodes
	Dst string `valid:"isGremlinExpr" yaml:"Dst"`
	// Metadata of the edges to create
	Metadata graph.Metadata `yaml:"Metadata"`
}

// GetName returns the resource name
func (e *EdgeRule) GetName() string {
	return "EdgeRule"
}

// Validate verifies the edge rule does not create invalid edges
func (e *EdgeRule) Validate() error {
	n1 := graph.CreateNode(graph.GenID(), nil, graph.TimeUTC(), "", "")
	n2 := graph.CreateNode(graph.GenID(), nil, graph.TimeUTC(), "", "")
	edge := graph.CreateEdge(graph.GenID(), n1, n2, e.Metadata, graph.TimeUTC(), "", "")
	return SchemaValidator.Validate("edge", edge)
}
