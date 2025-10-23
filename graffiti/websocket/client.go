package websocket

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	fmt "fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/safchain/insanelock"

	shttp "github.com/skydive-project/skydive/graffiti/http"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/graffiti/service"
)

const (
	maxMessageSize   = 0
	writeWait        = 10 * time.Second
	defaultQueueSize = 10000
)

// ConnState describes the connection state
type ConnState service.State

// ConnStatus describes the status of a WebSocket connection
type ConnStatus struct {
	ServiceType       service.Type
	ClientProtocol    Protocol
	Addr              string
	Port              int
	Host              string      `json:"-"`
	State             *ConnState  `json:"IsConnected"`
	URL               *url.URL    `json:"-"`
	Headers           http.Header `json:"-"`
	ConnectTime       time.Time
	RemoteHost        string       `json:",omitempty"`
	RemoteServiceType service.Type `json:",omitempty"`
}

// Store atomatically stores the state
func (s *ConnState) Store(state service.State) {
	(*service.State)(s).Store(state)
}

// Load atomatically loads and returns the state
func (s *ConnState) Load() service.State {
	return (*service.State)(s).Load()
}

// CompareAndSwap executes the compare-and-swap operation for a state
func (s *ConnState) CompareAndSwap(old, new service.State) bool {
	return atomic.CompareAndSwapInt64((*int64)(s), int64(old), int64(new))
}

// MarshalJSON marshal the connection state to JSON
func (s *ConnState) MarshalJSON() ([]byte, error) {
	switch service.State(*s) {
	case service.RunningState:
		return []byte("true"), nil
	case service.StoppedState:
		return []byte("false"), nil
	}
	return nil, fmt.Errorf("Invalid state: %d", s)
}

// UnmarshalJSON de-serialize a connection state
func (s *ConnState) UnmarshalJSON(b []byte) error {
	var state bool
	if err := json.Unmarshal(b, &state); err != nil {
		return err
	}

	if state {
		*s = ConnState(service.RunningState)
	} else {
		*s = ConnState(service.StoppedState)
	}

	return nil
}

// Message is the interface of a message to send over the wire
type Message interface {
	Bytes(protocol Protocol) ([]byte, error)
}

// RawMessage represents a raw message (array of bytes)
type RawMessage []byte

// Bytes returns the string representation of the raw message
func (m RawMessage) Bytes(protocol Protocol) ([]byte, error) {
	return m, nil
}

// Speaker is the interface for a websocket speaking client. It is used for outgoing
// or incoming connections.
type Speaker interface {
	GetStatus() ConnStatus
	GetHost() string
	GetAddrPort() (string, int)
	GetServiceType() service.Type
	GetClientProtocol() Protocol
	GetHeaders() http.Header
	GetURL() *url.URL
	IsConnected() bool
	SendMessage(Message) error
	SendRaw(r []byte) error
	Connect(context.Context) error
	Start()
	Stop()
	StopAndWait()
	Run()
	AddEventHandler(SpeakerEventHandler)
	GetRemoteHost() string
	GetRemoteServiceType() service.Type
}

// Conn is the connection object of a Speaker
type Conn struct {
	insanelock.RWMutex
	ConnStatus
	flush            chan struct{}
	send             chan []byte
	queueSize        int
	quit             chan bool
	wg               sync.WaitGroup
	conn             *websocket.Conn
	running          atomic.Value
	pingTicker       *time.Ticker // only used by incoming connections
	eventHandlers    []SpeakerEventHandler
	wsSpeaker        Speaker // speaker owning the connection
	writeCompression bool
	messageType      int
	logger           logging.Logger
}

// wsIncomingClient is only used internally to handle incoming client. It embeds a Conn.
type wsIncomingClient struct {
	*Conn
}

// Client is a outgoint client meaning a client connected to a remote websocket server.
// It embeds a Conn.
type Client struct {
	*Conn
	Path      string
	AuthOpts  *shttp.AuthenticationOpts
	TLSConfig *tls.Config
	Opts      ClientOpts
}

// ClientOpts defines some options that can be set when creating a new client
type ClientOpts struct {
	Protocol         Protocol
	AuthOpts         *shttp.AuthenticationOpts
	Headers          http.Header
	QueueSize        int
	WriteCompression bool
	TLSConfig        *tls.Config
	Logger           logging.Logger
}

// NewClientOpts returns a new client option set
func NewClientOpts() ClientOpts {
	return ClientOpts{
		Headers: http.Header{},
		Logger:  logging.GetLogger(),
	}
}

// SpeakerEventHandler is the interface to be implement by the client events listeners.
type SpeakerEventHandler interface {
	OnMessage(c Speaker, m Message)
	OnConnected(c Speaker) error
	OnDisconnected(c Speaker)
}

// DefaultSpeakerEventHandler implements stubs for the wsIncomingClientEventHandler interface
type DefaultSpeakerEventHandler struct {
}

// OnMessage is called when a message is received.
func (d *DefaultSpeakerEventHandler) OnMessage(c Speaker, m Message) {
}

// OnConnected is called when the connection is established.
func (d *DefaultSpeakerEventHandler) OnConnected(c Speaker) error {
	return nil
}

// OnDisconnected is called when the connection is closed or lost.
func (d *DefaultSpeakerEventHandler) OnDisconnected(c Speaker) {
}
