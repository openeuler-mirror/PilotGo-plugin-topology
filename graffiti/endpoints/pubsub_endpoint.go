package endpoints

import (
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/api/server"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph/traversal"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
	ws "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/websocket"
)

// PubSubEndpoint describes a WebSocket endpoint that can be used for both
// publishing and subscribing
type PubSubEndpoint struct {
	publisherEndpoint  *PublisherEndpoint
	subscriberEndpoint *SubscriberEndpoint
}

// NewPubSubEndpoint returns a new PubSub endpoint
func NewPubSubEndpoint(pool ws.StructSpeakerPool, validator server.Validator, g *graph.Graph, tr *traversal.GremlinTraversalParser, logger logging.Logger) *PubSubEndpoint {
	subscriberEndpoint := NewSubscriberEndpoint(pool, g, tr, logger)
	return &PubSubEndpoint{
		publisherEndpoint:  NewPublisherEndpoint(pool, g, validator, logger, subscriberEndpoint),
		subscriberEndpoint: subscriberEndpoint,
	}
}
