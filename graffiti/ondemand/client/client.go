package client

import (
	"encoding/json"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/api/rest"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"github.com/pmylund/go-cache"
	"github.com/safchain/insanelock"
	etcd "github.com/skydive-project/skydive/graffiti/etcd/client"
	ws "github.com/skydive-project/skydive/graffiti/websocket"
)

type OnDemandNodeResource struct {
	Node     *graph.Node
	Resource rest.Resource
}

// OnDemandClientHandler is the interface to be implemented by ondemand clients
type OnDemandClientHandler interface {
	ResourceName() string
	GetNodeResources(resource rest.Resource) []OnDemandNodeResource
	CheckState(n *graph.Node, resource rest.Resource) bool
	DecodeMessage(msg json.RawMessage) (rest.Resource, error)
	EncodeMessage(nodeID graph.Identifier, resource rest.Resource) (json.RawMessage, error)
}

// OnDemandClient describes an ondemand task client based on a websocket
type OnDemandClient struct {
	insanelock.RWMutex
	etcd.MasterElection
	graph.DefaultGraphListener
	graph                   *graph.Graph
	apiHandler              rest.WatchableHandler
	agentPool               ws.StructSpeakerPool
	subscriberPool          ws.StructSpeakerPool
	wsNamespace             string
	wsNotificationNamespace string
	resources               map[string]rest.Resource
	watcher                 rest.StoppableWatcher
	registeredNodes         map[graph.Identifier]map[string]bool
	deletedNodeCache        *cache.Cache
	checkForRegistration    *debouncer.Debouncer
	resourceName            string
	handler                 OnDemandClientHandler
}

type handlerNodeState struct {
	uuid    string
	started bool
}

type nodeTask struct {
	id       graph.Identifier
	host     string
	resource rest.Resource
}
