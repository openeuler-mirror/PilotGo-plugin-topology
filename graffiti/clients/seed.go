package clients

import (
	"fmt"
	"net/url"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/endpoints"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/filters"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/messages"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/service"
	ws "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/websocket"
)

// SeedService defines the seed service type
const SeedService service.Type = "seed"

// EventHandler is the interface to be implemented by event handler
type EventHandler interface {
	OnSynchronized()
}

// Seed is a service with its own graph. The seed synchronizes its
// graph by subscribing to the agent using WebSocket. It forwards
// all its graph events to the agent. A filter can be used to
// subscribe only to a part of the agent graph.
type Seed struct {
	*ws.Client
	ws.DefaultSpeakerEventHandler
	subscriber *Subscriber
	g          *graph.Graph
	logger     logging.Logger
}

// OnConnected websocket listener
func (s *Seed) OnConnected(c ws.Speaker) error {
	s.logger.Infof("connected to %s", c.GetHost())
	return s.subscriber.SendMessage(messages.NewStructMessage(messages.SyncRequestMsgType, messages.SyncRequestMsg{}))
}

// Start the seed
func (s *Seed) Start() {
	s.subscriber.Start()
	s.Client.Start()
}

// Stop the seed
func (s *Seed) Stop() {
	s.Client.Stop()
	s.subscriber.Stop()
}

// NewSeed returns a new seed
func NewSeed(g *graph.Graph, clientType service.Type, address, subscribeFilter, publishFilter string, wsOpts ws.ClientOpts, logger logging.Logger) (*Seed, error) {
	wsOpts.Headers.Add("X-Websocket-Namespace", messages.Namespace)

	if len(address) == 0 {
		address = "127.0.0.1:8081"
	}

	url, err := url.Parse("ws://" + address + "/ws/pubsub")
	if err != nil {
		return nil, fmt.Errorf("unable to parse the Address: %s, please check the configuration file", address)
	}

	wsOpts.Headers.Add("X-Gremlin-Filter", subscribeFilter)
	wsOpts.Headers.Add("X-Update-Policy", endpoints.PartialUpdates)

	pubsubClient := ws.NewClient(g.GetHost(), clientType, url, wsOpts)

	pool := ws.NewStructClientPool("pubsub", ws.PoolOpts{Logger: wsOpts.Logger})
	if err := pool.AddClient(pubsubClient); err != nil {
		return nil, fmt.Errorf("failed to add client: %w", err)
	}

	var metadataFilter *filters.Filter
	if publishFilter != "" {
		publishMetadata := graph.Metadata{}
		if _, err := graph.DefToMetadata(publishFilter, publishMetadata); err != nil {
			return nil, fmt.Errorf("failed to create publish filter: %w", err)
		}

		if metadataFilter, err = publishMetadata.Filter(); err != nil {
			return nil, fmt.Errorf("failed to create publish filter: %w", err)
		}
	}

	forwarder := NewForwarder(g, pool, metadataFilter, logger)
	subscriber := NewSubscriber(pubsubClient, g, wsOpts.Logger, forwarder)

	s := &Seed{
		Client:     pubsubClient,
		g:          g,
		subscriber: subscriber,
		logger:     wsOpts.Logger,
	}

	return s, nil
}
