package graph

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/safchain/insanelock"
	uuid "github.com/satori/go.uuid"

	"github.com/skydive-project/skydive/graffiti/filters"
	"github.com/skydive-project/skydive/graffiti/getter"
	"github.com/skydive-project/skydive/graffiti/logging"
)

const (
	maxEvents = 500
)

type graphEventType int

// Graph events
const (
	NodeUpdated graphEventType = iota + 1
	NodeAdded
	NodeDeleted
	EdgeUpdated
	EdgeAdded
	EdgeDeleted
)

const (
	// MetadataPrefix used to access user metadata
	MetadataPrefix = "Metadata."
)

// Identifier graph ID
type Identifier string

// EventListener describes the graph events interface mechanism
type EventListener interface {
	OnNodeUpdated(n *Node, ops []PartiallyUpdatedOp)
	OnNodeAdded(n *Node)
	OnNodeDeleted(n *Node)
	OnEdgeUpdated(e *Edge, ops []PartiallyUpdatedOp)
	OnEdgeAdded(e *Edge)
	OnEdgeDeleted(e *Edge)
}

type graphEvent struct {
	kind     graphEventType
	element  interface{}
	listener EventListener
	ops      []PartiallyUpdatedOp
}

type graphElement struct {
	ID        Identifier
	Metadata  Metadata
	Host      string
	Origin    string
	CreatedAt Time
	UpdatedAt Time
	DeletedAt Time `json:"DeletedAt,omitempty"`
	Revision  int64
}

// Node of the graph
type Node struct {
	graphElement `mapstructure:",squash"`
}

// Edge of the graph linked by a parent and a child
type Edge struct {
	graphElement `mapstructure:",squash"`
	Parent       Identifier
	Child        Identifier
}

// Graph errors
var (
	ErrElementNotFound = errors.New("Graph element not found")
	ErrNodeNotFound    = errors.New("Node not found")
	ErrEdgeNotFound    = errors.New("Edge not found")
	ErrParentNotFound  = errors.New("Parent node not found")
	ErrChildNotFound   = errors.New("Child node not found")
	ErrEdgeConflict    = errors.New("Edge ID conflict")
	ErrNodeConflict    = errors.New("Node ID conflict")
	ErrInternal        = errors.New("Internal backend error")
)

// Backend interface mechanism used as storage
type Backend interface {
	NodeAdded(n *Node) error
	NodeDeleted(n *Node) error
	GetNode(i Identifier, at Context) []*Node
	GetNodesFromIDs(i []Identifier, at Context) []*Node
	GetNodeEdges(n *Node, at Context, m ElementMatcher) []*Edge
	GetNodesEdges(n []*Node, at Context, m ElementMatcher) []*Edge

	EdgeAdded(e *Edge) error
	EdgeDeleted(e *Edge) error
	GetEdge(i Identifier, at Context) []*Edge
	GetEdgeNodes(e *Edge, at Context, parentMetadata, childMetadata ElementMatcher) ([]*Node, []*Node)

	MetadataUpdated(e interface{}) error

	GetNodes(t Context, m ElementMatcher) []*Node
	GetEdges(t Context, m ElementMatcher) []*Edge

	IsHistorySupported() bool
}

type PersistentBackendListener interface {
	OnStarted()
}

// PersistentBackend describes the interface of a persistent storage backend
type PersistentBackend interface {
	Backend

	AddListener(listener PersistentBackendListener)
	FlushElements(m ElementMatcher) error
	Sync(*Graph, *ElementFilter) error

	Start() error
	Stop()
}

// Context describes within time slice
type Context struct {
	TimeSlice *TimeSlice
	TimePoint bool
}

var liveContext = Context{TimePoint: true}

// Graph describes the graph object based on events and context mechanism
// An associated backend is used as storage
type Graph struct {
	insanelock.RWMutex
	eventHandler *EventHandler
	backend      Backend
	context      Context
	host         string
	origin       string
}

// Elements struct containing nodes and edges
type Elements struct {
	Nodes []*Node
	Edges []*Edge
}

// MetadataDecoder defines a json rawmessage decoder which has to return a object
// implementing the getter interface
type MetadataDecoder func(raw json.RawMessage) (getter.Getter, error)

var (
	// NodeMetadataDecoders is a map that owns special type metadata decoder
	NodeMetadataDecoders = make(map[string]MetadataDecoder)
	// EdgeMetadataDecoders is a map that owns special type metadata decoder
	EdgeMetadataDecoders = make(map[string]MetadataDecoder)
)

// DefaultGraphListener default implementation of a graph listener, can be used when not implementing
// the whole set of callbacks
type DefaultGraphListener struct{}

// String return the name of a graphEventType
func (g graphEventType) String() string {
	return []string{
		"", "NodeUpdated", "NodeAdded", "NodeDeleted", "EdgeUpdated", "EdgeAdded", "EdgeDeleted",
	}[g]
}

// OnNodeUpdated event
func (c *DefaultGraphListener) OnNodeUpdated(n *Node, ops []PartiallyUpdatedOp) {
}

// OnNodeAdded event
func (c *DefaultGraphListener) OnNodeAdded(n *Node) {
}

// OnNodeDeleted event
func (c *DefaultGraphListener) OnNodeDeleted(n *Node) {
}

// OnEdgeUpdated event
func (c *DefaultGraphListener) OnEdgeUpdated(e *Edge, ops []PartiallyUpdatedOp) {
}

// OnEdgeAdded event
func (c *DefaultGraphListener) OnEdgeAdded(e *Edge) {
}

// OnEdgeDeleted event
func (c *DefaultGraphListener) OnEdgeDeleted(e *Edge) {
}

// ListenerHandler describes an other that manages a set of event listeners
type ListenerHandler interface {
	AddEventListener(l EventListener)
	RemoveEventListener(l EventListener)
}

// EventHandler describes an object that notifies listeners with graph events
type EventHandler struct {
	insanelock.RWMutex
	eventListeners       []EventListener
	eventChan            chan graphEvent
	eventConsumed        bool
	currentEventListener EventListener
}

// PartiallyUpdatedOpType operation type add/del
type PartiallyUpdatedOpType int

const (
	// PartiallyUpdatedAddOpType add metadata
	PartiallyUpdatedAddOpType PartiallyUpdatedOpType = iota + 1
	// PartiallyUpdatedDelOpType del metadata
	PartiallyUpdatedDelOpType
)

// PartiallyUpdatedOp describes a way to update partially node or edge
type PartiallyUpdatedOp struct {
	Type  PartiallyUpdatedOpType
	Key   string
	Value interface{}
}

// NodePartiallyUpdated partial updates of a node
type NodePartiallyUpdated struct {
	Node *Node
	Ops  []PartiallyUpdatedOp
}

// EdgePartiallyUpdated partial updates of a edge
type EdgePartiallyUpdated struct {
	Edge *Edge
	Ops  []PartiallyUpdatedOp
}
