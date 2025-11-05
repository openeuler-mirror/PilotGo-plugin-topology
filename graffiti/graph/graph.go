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

func (g *EventHandler) notifyListeners(ge graphEvent) {
	// notify only once per listener as if more than once we are in a recursion
	// and we wont to notify a listener which generated a graph element
	g.RLock()
	defer g.RUnlock()
	for _, g.currentEventListener = range g.eventListeners {
		// do not notify the listener which generated the event
		if g.currentEventListener == ge.listener {
			continue
		}

		switch ge.kind {
		case NodeAdded:
			g.currentEventListener.OnNodeAdded(ge.element.(*Node))
		case NodeUpdated:
			g.currentEventListener.OnNodeUpdated(ge.element.(*Node), ge.ops)
		case NodeDeleted:
			g.currentEventListener.OnNodeDeleted(ge.element.(*Node))
		case EdgeAdded:
			g.currentEventListener.OnEdgeAdded(ge.element.(*Edge))
		case EdgeUpdated:
			g.currentEventListener.OnEdgeUpdated(ge.element.(*Edge), ge.ops)
		case EdgeDeleted:
			g.currentEventListener.OnEdgeDeleted(ge.element.(*Edge))
		}
	}
}

// NotifyEvent notifies all the listeners of an event. NotifyEvent
// makes sure that we don't enter a notify endless loop.
func (g *EventHandler) NotifyEvent(kind graphEventType, element interface{}, ops ...PartiallyUpdatedOp) {
	// push event to chan so that nested notification will be sent in the
	// right order. Associate the event with the current event listener so
	// we can avoid loop by not triggering event for the current listener.
	ge := graphEvent{kind: kind, element: element, ops: ops}
	ge.listener = g.currentEventListener
	select {
	case g.eventChan <- ge:
	default:
		logging.GetLogger().Errorf("maxEvents reached. Ignoring event '%+v' for element '%+v'", kind, element)
	}

	// already a consumer no need to run another consumer
	if g.eventConsumed {
		return
	}
	g.eventConsumed = true

	for len(g.eventChan) > 0 {
		ge = <-g.eventChan
		g.notifyListeners(ge)
	}
	g.currentEventListener = nil
	g.eventConsumed = false
}

// AddEventListener subscribe a new graph listener
func (g *EventHandler) AddEventListener(l EventListener) {
	g.Lock()
	defer g.Unlock()

	g.eventListeners = append(g.eventListeners, l)
}

// RemoveEventListener unsubscribe a graph listener
func (g *EventHandler) RemoveEventListener(l EventListener) {
	g.Lock()
	defer g.Unlock()

	for i, el := range g.eventListeners {
		if l == el {
			g.eventListeners = append(g.eventListeners[:i], g.eventListeners[i+1:]...)
			break
		}
	}
}

// NewEventHandler instantiate a new event handler
func NewEventHandler(maxEvents int) *EventHandler {
	return &EventHandler{
		eventChan: make(chan graphEvent, maxEvents),
	}
}

// GenID helper generate a node Identifier
func GenID(s ...string) Identifier {
	if len(s) > 0 {
		return Identifier(uuid.NewV5(uuid.NamespaceOID, strings.Join(s, "/")).String())
	}

	return Identifier(uuid.NewV4().String())
}

func (e *graphElement) GetFieldBool(field string) (_ bool, err error) {
	return e.Metadata.GetFieldBool(strings.TrimPrefix(field, MetadataPrefix))
}

func (e *graphElement) GetFieldInt64(field string) (_ int64, err error) {
	switch field {
	case "@CreatedAt":
		return e.CreatedAt.UnixMilli(), nil
	case "@UpdatedAt":
		return e.UpdatedAt.UnixMilli(), nil
	case "@DeletedAt":
		return e.DeletedAt.UnixMilli(), nil
	case "@Revision":
		return e.Revision, nil
	default:
		return e.Metadata.GetFieldInt64(strings.TrimPrefix(field, MetadataPrefix))
	}
}

func (e *graphElement) GetFieldString(field string) (_ string, err error) {
	switch field {
	case "@ID":
		return string(e.ID), nil
	case "@Host":
		return e.Host, nil
	case "@Origin":
		return e.Origin, nil
	default:
		return e.Metadata.GetFieldString(strings.TrimPrefix(field, MetadataPrefix))
	}
}

func (e *graphElement) GetField(field string) (interface{}, error) {
	if i, err := e.GetFieldInt64(field); err == nil {
		return i, nil
	}

	if s, err := e.GetFieldString(field); err == nil {
		return s, nil
	}

	return e.Metadata.GetField(strings.TrimPrefix(field, MetadataPrefix))
}

func (e *graphElement) GetFields(names []string) (interface{}, error) {
	values := make(map[string]interface{})

	for _, name := range names {
		if v, err := e.GetField(name); err == nil {
			values[name] = v
		}
	}

	if len(values) == 0 {
		return nil, getter.ErrFieldNotFound
	}

	return values, nil
}

var graphElementKeys = map[string]bool{"@ID": false, "@Host": false, "@Origin": false, "@CreatedAt": false, "@UpdatedAt": false, "@DeletedAt": false, "@Revision": false}

func (e *graphElement) GetFieldKeys() []string {
	keys := make([]string, len(graphElementKeys))
	i := 0
	for key := range graphElementKeys {
		keys[i] = key
		i++
	}
	return append(keys, e.Metadata.GetFieldKeys()...)
}

func (e *graphElement) MatchBool(field string, predicate getter.BoolPredicate) bool {
	if index := strings.Index(field, "."); index != -1 {
		first := field[index+1:]
		if v, found := e.Metadata[first]; found {
			if getter, found := v.(getter.Getter); found {
				return getter.MatchBool(field[index+1:], predicate)
			}
		}
	}

	field = strings.TrimPrefix(field, MetadataPrefix)

	return e.Metadata.MatchBool(field, predicate)
}

func (e *graphElement) MatchInt64(field string, predicate getter.Int64Predicate) bool {
	if _, found := graphElementKeys[field]; found {
		if n, err := e.GetFieldInt64(field); err == nil {
			return predicate(n)
		}
	}

	field = strings.TrimPrefix(field, MetadataPrefix)

	if index := strings.Index(field, "."); index != -1 {
		first := field[index+1:]
		if v, found := e.Metadata[first]; found {
			if getter, found := v.(getter.Getter); found {
				return getter.MatchInt64(field[index+1:], predicate)
			}
		}
	}

	return e.Metadata.MatchInt64(field, predicate)
}

func (e *graphElement) MatchString(field string, predicate getter.StringPredicate) bool {
	if _, found := graphElementKeys[field]; found {
		if s, err := e.GetFieldString(field); err == nil {
			return predicate(s)
		}
	}

	field = strings.TrimPrefix(field, MetadataPrefix)

	if index := strings.Index(field, "."); index != -1 {
		first := field[index+1:]
		if v, found := e.Metadata[first]; found {
			if getter, found := v.(getter.Getter); found {
				return getter.MatchString(field[index+1:], predicate)
			}
		}
	}

	return e.Metadata.MatchString(field, predicate)
}

func (e *graphElement) GetFieldStringList(key string) ([]string, error) {
	value, err := e.GetField(key)
	if err != nil {
		return nil, err
	}

	switch value := value.(type) {
	case []interface{}:
		var strings []string
		for _, s := range value {
			if s, ok := s.(string); ok {
				strings = append(strings, s)
			} else {
				return nil, errors.New("one array entry could not be converted to string")
			}
		}
		return strings, nil
	case []string:
		return value, nil
	default:
		return nil, getter.ErrFieldNotFound
	}
}

// MatchMetadata returns whether a graph element matches with the provided filter or metadata
func (e *graphElement) MatchMetadata(f ElementMatcher) bool {
	if f == nil {
		return true
	}
	return f.Match(e)
}

func normalizeNumbers(raw interface{}) interface{} {
	switch t := raw.(type) {
	case float64:
		// float equal to its int version then convert it to int64
		i := int64(t)
		if float64(i) == t {
			return i
		}
	case []interface{}:
		for i, obj := range t {
			t[i] = normalizeNumbers(obj)
		}
	case map[string]interface{}:
		for k, v := range t {
			t[k] = normalizeNumbers(v)
		}
		return t
	}
	return raw
}

// normalize the graph element after partial deserialization
func (e *graphElement) normalize(metadata map[string]json.RawMessage, decoders map[string]MetadataDecoder) error {
	// sanity checks
	if e.ID == "" {
		return errors.New("No ID found for graph element")
	}

	e.Metadata = make(Metadata)

	if e.CreatedAt.IsZero() {
		e.CreatedAt = TimeUTC()
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = TimeUTC()
	}

	for k, r := range metadata {
		if len(r) == 0 {
			continue
		}

		if decoder, ok := decoders[k]; ok {
			v, err := decoder(r)
			if err != nil {
				return err
			}
			e.Metadata[k] = v
			continue
		}

		var i interface{}
		if err := json.Unmarshal(r, &i); err != nil {
			return err
		}

		e.Metadata[k] = normalizeNumbers(i)
	}

	return nil
}
