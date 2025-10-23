package websocket

import (
	"encoding/json"
	"errors"
	fmt "fmt"
	"net/http"
	"time"

	auth "github.com/abbot/go-http-auth"
	proto "github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/safchain/insanelock"
	uuid "github.com/satori/go.uuid"

	"github.com/skydive-project/skydive/graffiti/logging"
)

// Protocol used to transport messages
type Protocol string

const (
	// WildcardNamespace is the namespace used as wildcard. It is used by listeners to filter callbacks.
	WildcardNamespace = "*"

	// RawProtocol is used for raw messages
	RawProtocol Protocol = "raw"
	// ProtobufProtocol is used for protobuf encoded messages
	ProtobufProtocol Protocol = "protobuf"
	// JSONProtocol is used for JSON encoded messages
	JSONProtocol Protocol = "json"
)

// DefaultRequestTimeout default timeout used for Request/Reply JSON message.
var DefaultRequestTimeout = 10 * time.Second

// ProtobufObject defines an object that can be serialized in protobuf
type ProtobufObject interface {
	proto.Marshaler
	proto.Message
}

type structMessageState struct {
	value interface{}

	// keep a cached version to avoid to serialize multiple time during
	// the call flow
	jsonCache     []byte
	protobufCache []byte
}

func (p *Protocol) parse(s string) error {
	switch s {
	case "", "json":
		*p = JSONProtocol
	case "protobuf":
		*p = ProtobufProtocol
	default:
		return errors.New("protocol not supported")
	}
	return nil
}

func (p *Protocol) String() string {
	return string(*p)
}

// Debug representation of the struct StructMessage
func (g *StructMessage) Debug() string {
	return fmt.Sprintf("Namespace %s Type %s UUID %s Status %d Obj (%d bytes)",
		g.Namespace, g.Type, g.UUID, g.Status, len(g.Obj))
}

func (g *StructMessage) protobufBytes() ([]byte, error) {
	if len(g.XXX_state.protobufCache) > 0 {
		return g.XXX_state.protobufCache, nil
	}

	obj, ok := g.XXX_state.value.(ProtobufObject)
	if ok {
		g.Format = Format_Protobuf

		msgObj, err := obj.Marshal()
		if err != nil {
			return nil, err
		}
		g.Obj = msgObj
	} else {
		// fallback to json
		g.Format = Format_Json

		msgObj, err := json.Marshal(g.XXX_state.value)
		if err != nil {
			return nil, err
		}
		g.Obj = msgObj
	}

	msgBytes, err := g.Marshal()
	if err != nil {
		return nil, err
	}

	g.XXX_state.protobufCache = msgBytes

	return msgBytes, nil
}

func (g *StructMessage) jsonBytes() ([]byte, error) {
	if len(g.XXX_state.jsonCache) > 0 {
		return g.XXX_state.jsonCache, nil
	}

	m := struct {
		*StructMessage
		Obj interface{}
	}{
		StructMessage: g,
		Obj:           g.XXX_state.value,
	}

	msgBytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	g.XXX_state.jsonCache = msgBytes

	return msgBytes, nil
}

// Bytes implements the message interface
func (g *StructMessage) Bytes(protocol Protocol) ([]byte, error) {
	if protocol == ProtobufProtocol {
		return g.protobufBytes()
	}
	return g.jsonBytes()
}

// UnmarshalJSON custom unmarshal
func (g *StructMessage) UnmarshalJSON(b []byte) error {
	m := struct {
		Namespace string
		Type      string
		UUID      string
		Status    int64
		Obj       json.RawMessage
	}{}

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	g.Namespace = m.Namespace
	g.Type = m.Type
	g.UUID = m.UUID
	g.Status = m.Status
	g.Obj = []byte(m.Obj)

	return nil
}

func (g *StructMessage) unmarshalByProtocol(b []byte, protocol Protocol) error {
	if protocol == ProtobufProtocol {
		if err := g.Unmarshal(b); err != nil {
			return fmt.Errorf("Error while decoding Protobuf StructMessage %s", err)
		}
	} else {
		if err := g.UnmarshalJSON(b); err != nil {
			return fmt.Errorf("Error while decoding JSON StructMessage %s", err)
		}
	}

	return nil
}

// NewStructMessage creates a new StructMessage with the given namespace, type, value
// and optionally the UUID.
func NewStructMessage(ns string, tp string, v interface{}, uuids ...string) *StructMessage {
	var u string
	if len(uuids) != 0 {
		u = uuids[0]
	} else {
		u = uuid.NewV4().String()
	}

	msg := &StructMessage{
		Namespace: ns,
		Type:      tp,
		UUID:      u,
		Status:    int64(http.StatusOK),
	}
	msg.XXX_state.value = v

	return msg
}

// Reply returns a reply message with the given value, type and status.
// Basically it return a new StructMessage with the correct Namespace and UUID.
func (g *StructMessage) Reply(v interface{}, kind string, status int) *StructMessage {
	msg := &StructMessage{
		Namespace: g.Namespace,
		Type:      kind,
		UUID:      g.UUID,
		Status:    int64(status),
	}
	msg.XXX_state.value = v

	return msg
}

// SpeakerStructMessageDispatcher interface is used to dispatch OnStructMessage events.
type SpeakerStructMessageDispatcher interface {
	AddStructMessageHandler(h SpeakerStructMessageHandler, namespaces []string)
}

type structSpeakerEventDispatcher struct {
	eventHandlersLock insanelock.RWMutex
	nsEventHandlers   map[string][]SpeakerStructMessageHandler
}

func newStructSpeakerEventDispatcher() *structSpeakerEventDispatcher {
	return &structSpeakerEventDispatcher{
		nsEventHandlers: make(map[string][]SpeakerStructMessageHandler),
	}
}

// AddStructMessageHandler adds a new listener for Struct messages.
func (a *structSpeakerEventDispatcher) AddStructMessageHandler(h SpeakerStructMessageHandler, namespaces []string) {
	a.eventHandlersLock.Lock()
	// add this handler per namespace
	for _, ns := range namespaces {
		if _, ok := a.nsEventHandlers[ns]; !ok {
			a.nsEventHandlers[ns] = []SpeakerStructMessageHandler{h}
		} else {
			a.nsEventHandlers[ns] = append(a.nsEventHandlers[ns], h)
		}
	}
	a.eventHandlersLock.Unlock()
}

func (a *structSpeakerEventDispatcher) dispatchMessage(c *StructSpeaker, m *StructMessage) {
	// check whether it is a reply
	if c.onReply(m) {
		return
	}

	a.eventHandlersLock.RLock()
	var handlers []SpeakerStructMessageHandler
	handlers = append(handlers, a.nsEventHandlers[m.Namespace]...)
	handlers = append(handlers, a.nsEventHandlers[WildcardNamespace]...)
	a.eventHandlersLock.RUnlock()

	for _, h := range handlers {
		h.OnStructMessage(c, m)
	}
}

// OnDisconnected is implemented here to avoid infinite loop since the default
// implementation is triggering OnDisconnected too.
func (a *structSpeakerEventDispatcher) OnDisconnected(c Speaker) {
}

// OnConnected is implemented here to avoid infinite loop since the default
// implementation is triggering OnDisconnected too.
func (a *structSpeakerEventDispatcher) OnConnected(c Speaker) error {
	return nil
}

type structSpeakerPoolEventDispatcher struct {
	dispatcher *structSpeakerEventDispatcher
	pool       SpeakerPool
}

// AddStructMessageHandler adds a new listener for Struct messages.
func (d *structSpeakerPoolEventDispatcher) AddStructMessageHandler(h SpeakerStructMessageHandler, namespaces []string) {
	d.dispatcher.AddStructMessageHandler(h, namespaces)
	for _, client := range d.pool.GetSpeakers() {
		client.(*StructSpeaker).AddStructMessageHandler(h, namespaces)
	}
}

func (d *structSpeakerPoolEventDispatcher) AddStructSpeaker(c *StructSpeaker) {
	d.dispatcher.eventHandlersLock.RLock()
	for ns, handlers := range d.dispatcher.nsEventHandlers {
		for _, handler := range handlers {
			c.AddStructMessageHandler(handler, []string{ns})
		}
	}
	d.dispatcher.eventHandlersLock.RUnlock()
}

func newStructSpeakerPoolEventDispatcher(pool SpeakerPool) *structSpeakerPoolEventDispatcher {
	return &structSpeakerPoolEventDispatcher{
		dispatcher: newStructSpeakerEventDispatcher(),
		pool:       pool,
	}
}

