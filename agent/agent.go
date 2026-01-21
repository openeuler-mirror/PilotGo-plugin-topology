package agent

import (
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/ondemand/server"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/pod"
	"github.com/skydive-project/skydive/flow/client"
	shttp "github.com/skydive-project/skydive/graffiti/http"
	ws "github.com/skydive-project/skydive/graffiti/websocket"
	"github.com/skydive-project/skydive/probe"
	"github.com/skydive-project/skydive/topology"
)

// Agent object started on each hosts/namespaces
type Agent struct {
	ws.DefaultSpeakerEventHandler
	pod                 *pod.Pod
	topologyProbeBundle *probe.Bundle
	flowProbeBundle     *probe.Bundle
	flowClientPool      *client.FlowClientPool
	onDemandProbeServer *server.OnDemandServer
	onDemandPIServer    *server.OnDemandServer
	httpServer          *shttp.Server
	tidMapper           *topology.TIDMapper
}

// Status agent object
//
// # Status describes the status of an agent
//
// swagger:model AgentStatus
// easyjson:json
type Status struct {
	Clients        map[string]ws.ConnStatus
	Analyzers      map[string]pod.ConnStatus
	TopologyProbes map[string]interface{}
	FlowProbes     []string
}
