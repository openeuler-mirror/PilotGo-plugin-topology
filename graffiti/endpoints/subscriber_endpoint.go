package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/safchain/insanelock"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph/traversal"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/messages"
	ws "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/websocket"
)

// UpdatePolicy specifies whether we should send partial or full updates
type UpdatePolicy = string

// Update policies
const (
	FullUpdates    UpdatePolicy = "full"
	PartialUpdates              = "partial"
)

type subscriber struct {
	ws.Speaker
	lastGraph     *graph.Graph
	gremlinFilter string
	ts            *traversal.GremlinTraversalSequence
	updatePolicy  UpdatePolicy
	inhibit       atomic.Value
}

func (s *subscriber) getSubGraph(g *graph.Graph, lockGraph bool) (*graph.Graph, error) {
	res, err := s.ts.Exec(g, lockGraph)
	if err != nil {
		return nil, err
	}

	tv, ok := res.(*traversal.GraphTraversal)
	if !ok {
		return nil, fmt.Errorf("Gremlin query '%s' did not return a graph", s.gremlinFilter)
	}

	return tv.Graph, nil
}

func (s *subscriber) resetFilter() {
	s.gremlinFilter = ""
	s.ts = nil
	s.lastGraph = nil
}

// SubscriberEndpoint sends all the modifications to its subscribers.
type SubscriberEndpoint struct {
	insanelock.RWMutex
	ws.DefaultSpeakerEventHandler
	pool          ws.StructSpeakerPool
	graph         *graph.Graph
	wg            sync.WaitGroup
	gremlinParser *traversal.GremlinTraversalParser
	subscribers   map[ws.Speaker]*subscriber
	logger        logging.Logger
	inhib         atomic.Value
}

func (t *SubscriberEndpoint) newSubscriber(speaker ws.Speaker, gremlinFilter string, lockGraph bool) (s *subscriber, err error) {
	s = &subscriber{Speaker: speaker, gremlinFilter: gremlinFilter}

	if gremlinFilter != "" {
		s.ts, err = t.gremlinParser.Parse(strings.NewReader(gremlinFilter))
		if err != nil {
			return nil, fmt.Errorf("Invalid Gremlin filter '%s' for client %s", gremlinFilter, speaker.GetRemoteHost())
		}

		if s.lastGraph, err = s.getSubGraph(t.graph, lockGraph); err != nil {
			return nil, err
		}
	} else {
		s.lastGraph = t.graph
	}

	return s, nil
}

// OnConnected called when a subscriber got connected.
func (t *SubscriberEndpoint) OnConnected(c ws.Speaker) error {
	gremlinFilter := c.GetHeaders().Get("X-Gremlin-Filter")
	if gremlinFilter == "" {
		gremlinFilter = c.GetURL().Query().Get("x-gremlin-filter")
	}

	subscriber, err := t.newSubscriber(c, gremlinFilter, true)
	if err != nil {
		t.logger.Error(err)
		return err
	}

	updatePolicy := c.GetHeaders().Get("X-Update-Policy")
	if updatePolicy == "" {
		updatePolicy = c.GetURL().Query().Get("x-update-policy")
	}

	switch strings.ToLower(updatePolicy) {
	case FullUpdates, PartialUpdates:
	case "":
		updatePolicy = FullUpdates
	default:
		err = fmt.Errorf("invalid update policy '%s'", subscriber.updatePolicy)
		t.logger.Error(err)
		return err
	}

	subscriber.updatePolicy = updatePolicy

	t.logger.Infof("Client %s subscribed with filter '%s' and update policy '%s'", c.GetRemoteHost(), gremlinFilter, updatePolicy)
	t.Lock()
	t.subscribers[c] = subscriber
	t.Unlock()

	return nil
}

// OnDisconnected called when a subscriber got disconnected.
func (t *SubscriberEndpoint) OnDisconnected(c ws.Speaker) {
	t.Lock()
	delete(t.subscribers, c)
	t.Unlock()
}

// OnStructMessage is triggered when receiving a message from a subscriber.
// It only responds to SyncRequestMsgType messages
func (t *SubscriberEndpoint) OnStructMessage(c ws.Speaker, msg *ws.StructMessage) {
	t.logger.Debug("Graph: Received a message from client %s", c.GetRemoteHost())
	msgType, obj, err := messages.UnmarshalMessage(msg)
	if err != nil {
		t.logger.Errorf("Graph: Unable to parse the event %v: %s", msg, err)
		return
	}

	// this kind of message usually comes from external clients like the WebUI
	if msgType == messages.SyncRequestMsgType {
		t.graph.RLock()
		defer t.graph.RUnlock()

		t.logger.Debug("Graph: Received a SyncRequest message from client %s", c.GetRemoteHost())
		syncMsg, status := obj.(*messages.SyncRequestMsg), http.StatusOK
		result, err := t.graph.CloneWithContext(syncMsg.Context)
		if err != nil {
			t.logger.Errorf("unable to get a graph with context %+v: %s", syncMsg, err)
			reply := msg.Reply(nil, messages.SyncReplyMsgType, http.StatusBadRequest)
			c.SendMessage(reply)
			return
		}

		host := c.GetRemoteHost()

		if syncMsg.GremlinFilter != nil {
			// filter reset
			if *syncMsg.GremlinFilter == "" {
				t.Lock()
				t.subscribers[c].resetFilter()
				t.Unlock()
			} else {
				subscriber, err := t.newSubscriber(c, *syncMsg.GremlinFilter, false)
				if err != nil {
					t.logger.Error(err)

					reply := msg.Reply(err.Error(), messages.SyncReplyMsgType, http.StatusBadRequest)
					c.SendMessage(reply)

					return
				}

				t.logger.Infof("Client %s requested subscription with filter %s", host, *syncMsg.GremlinFilter)
				result = subscriber.lastGraph

				t.Lock()
				t.subscribers[c] = subscriber
				t.Unlock()
			}
		} else {
			t.RLock()
			subscriber := t.subscribers[c]
			t.RUnlock()

			if subscriber != nil {
				result = subscriber.lastGraph
			}
		}

		reply := msg.Reply(result, messages.SyncReplyMsgType, status)
		replyJson, _ := json.Marshal(reply)
		t.logger.Debug("Graph: Sending a SyncReply message to client:  %s", string(replyJson))
		resultJson, _ := json.Marshal(result)
		t.logger.Debug("Graph: Sending a SyncReply message to client:  %s", string(resultJson))
		c.SendMessage(reply)

		return
	}
}
