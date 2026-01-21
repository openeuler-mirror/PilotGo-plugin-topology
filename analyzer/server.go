package analyzer

import (
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/hub"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/ondemand/client"
	"github.com/skydive-project/skydive/flow/server"
	"github.com/skydive-project/skydive/flow/storage"
	etcdclient "github.com/skydive-project/skydive/graffiti/etcd/client"
	ws "github.com/skydive-project/skydive/graffiti/websocket"
	"github.com/skydive-project/skydive/probe"
	usertopology "github.com/skydive-project/skydive/topology/enhancers"
	"github.com/skydive-project/skydive/ui"
)

const workflowAssetDir = "statics/workflows"

// Status analyzer object
//
// # Status describes the status of an analyzer
//
// swagger:model AnalyzerStatus
// easyjson:json
type Status struct {
	Agents      map[string]ws.ConnStatus
	Peers       hub.PeersStatus
	Publishers  map[string]ws.ConnStatus
	Subscribers map[string]ws.ConnStatus
	Alerts      hub.ElectionStatus
	Captures    hub.ElectionStatus
	Probes      map[string]interface{}
}

// Server describes an Analyzer servers mechanism like http, websocket, topology, ondemand probes, ...
type Server struct {
	uiServer        *ui.Server
	hub             *hub.Hub
	onDemandClient  *client.OnDemandClient
	piClient        *client.OnDemandClient
	topologyManager *usertopology.TopologyManager
	flowServer      *server.FlowServer
	probeBundle     *probe.Bundle
	graphStorage    graph.PersistentBackend
	flowStorage     storage.Storage
	etcdClient      *etcdclient.Client
}

// GetStatus returns the status of an analyzer
func (s *Server) GetStatus() interface{} {
	hubStatus := s.hub.GetStatus().(*hub.Status)

	return &Status{
		Agents:      hubStatus.Pods,
		Peers:       hubStatus.Peers,
		Publishers:  hubStatus.Publishers,
		Subscribers: hubStatus.Subscribers,
		Alerts:      hubStatus.Alerts,
		Captures:    hub.ElectionStatus{IsMaster: s.onDemandClient.IsMaster()},
		Probes:      s.probeBundle.GetStatus(),
	}
}
