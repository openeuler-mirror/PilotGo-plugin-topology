package endpoints

import (
	"errors"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/safchain/insanelock"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/http"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/messages"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/service"
	ws "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/websocket"
)

var (
	errAlreadyConnected  = errors.New("already connected")
	errConnectedToItself = errors.New("connected to itself")
)

// ReplicatorPeer is a remote connection to another Graph server. Only modification
// of the local Graph made either by the local server, by an agent message or by an external
// client will be forwarded to the peer.
type ReplicatorPeer struct {
	ws.DefaultSpeakerEventHandler
	URL       *url.URL
	Graph     *graph.Graph
	opts      *ws.ClientOpts
	wsspeaker ws.Speaker
	endpoint  *ReplicationEndpoint
}

type peerState struct {
	cnt int
}

// ReplicationEndpoint serves the local Graph and send local modification to its peers.
type ReplicationEndpoint struct {
	insanelock.RWMutex
	ws.DefaultSpeakerEventHandler
	in           ws.StructSpeakerPool
	out          *ws.StructClientPool
	peerStates   map[string]*peerState
	candidates   []*ReplicatorPeer
	Graph        *graph.Graph
	cached       *graph.CachedBackend
	replicateMsg atomic.Value
	wg           sync.WaitGroup
	logger       logging.Logger
}

// OnConnected is called when the peer gets connected then the whole graph
// is send to initialize it.
func (p *ReplicatorPeer) OnConnected(c ws.Speaker) error {
	p.endpoint.Lock()
	defer p.endpoint.Unlock()

	host := c.GetRemoteHost()
	if c.GetHost() == host {
		p.endpoint.logger.Debugf("Disconnecting from %s since it's me", p.URL.String())
		return errConnectedToItself
	}

	p.Graph.RLock()
	defer p.Graph.RUnlock()

	state, ok := p.endpoint.peerStates[host]
	if !ok {
		state = &peerState{}
	}

	// disconnect as can be connected to the same host from different addresses.
	if state.cnt > 0 {
		p.endpoint.logger.Debugf("Disconnecting from %s as already connected through %s", p.URL.String(), c.GetURL().String())
		return errAlreadyConnected
	}

	p.endpoint.logger.Infof("Connected to peer %s for replication", host)

	p.endpoint.peerStates[host] = state
	state.cnt++

	msg := &messages.SyncMsg{
		Elements: p.Graph.Elements(),
	}

	if err := p.wsspeaker.SendMessage(messages.NewStructMessage(messages.SyncMsgType, msg)); err != nil {
		return err
	}

	return p.endpoint.out.AddClient(c)
}

// OnDisconnected is called when the peer gets disconnected
func (p *ReplicatorPeer) OnDisconnected(c ws.Speaker) {
	p.endpoint.Lock()
	defer p.endpoint.Unlock()

	host := c.GetRemoteHost()
	if host == c.GetHost() {
		return
	}

	state := p.endpoint.peerStates[host]
	state.cnt--
	if state.cnt > 0 {
		return
	}

	origin := graph.ClientOrigin(c)
	if p.Graph.GetOrigin() == origin {
		return
	}

	p.endpoint.out.RemoveClient(c)
	delete(p.endpoint.peerStates, host)

	p.Graph.Lock()
	graph.DelSubGraphOfOrigin(p.Graph, origin)
	p.Graph.Unlock()
}

func (p *ReplicatorPeer) connect(wg *sync.WaitGroup) {
	defer wg.Done()

	p.endpoint.logger.Infof("Connecting to peer: %s", p.URL.String())
	serviceType := service.Type(strings.SplitN(p.Graph.GetOrigin(), ".", 2)[0])
	wsClient := ws.NewClient(p.Graph.GetHost(), serviceType, p.URL, *p.opts)

	structClient := wsClient.UpgradeToStructSpeaker()
	// will trigger shttp.SpeakerEventHandler, so OnConnected
	structClient.AddEventHandler(p)

	// subscribe to the graph messages
	structClient.AddStructMessageHandler(p.endpoint, []string{messages.Namespace})

	p.wsspeaker = structClient
	p.wsspeaker.Start()
}

func (p *ReplicatorPeer) disconnect() {
	if p.wsspeaker != nil {
		p.wsspeaker.Stop()
	}
}

func (t *ReplicationEndpoint) addCandidate(addr string, port int, opts *ws.ClientOpts) *ReplicatorPeer {
	url, _ := http.MakeURL("ws", addr, port, "/ws/replication", opts.TLSConfig != nil)
	peer := &ReplicatorPeer{
		URL:      url,
		Graph:    t.Graph,
		opts:     opts,
		endpoint: t,
	}

	t.candidates = append(t.candidates, peer)
	return peer
}

// ConnectPeers starts a goroutine connecting all the peers.
func (t *ReplicationEndpoint) ConnectPeers() {
	t.RLock()
	defer t.RUnlock()

	for _, candidate := range t.candidates {
		t.wg.Add(1)
		go candidate.connect(&t.wg)
	}
}

// DisconnectPeers disconnects all the peers and wait until all disconnected.
func (t *ReplicationEndpoint) DisconnectPeers() {
	t.RLock()
	defer t.RUnlock()

	for _, candidate := range t.candidates {
		candidate.disconnect()
	}
	t.wg.Wait()
}
