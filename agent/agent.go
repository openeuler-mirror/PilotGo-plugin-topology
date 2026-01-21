package agent

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	api "github.com/skydive-project/skydive/api/server"
	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/flow/client"
	"github.com/skydive-project/skydive/graffiti/graph"
	shttp "github.com/skydive-project/skydive/graffiti/http"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/graffiti/ondemand/server"
	"github.com/skydive-project/skydive/graffiti/pod"
	ws "github.com/skydive-project/skydive/graffiti/websocket"
	"github.com/skydive-project/skydive/probe"
	"github.com/skydive-project/skydive/topology"
	"github.com/skydive-project/skydive/validator"
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

// GetStatus returns the status of an agent
func (a *Agent) GetStatus() interface{} {
	podStatus := a.pod.GetStatus().(*pod.Status)

	return &Status{
		Clients:        podStatus.Subscribers,
		Analyzers:      podStatus.Hubs,
		TopologyProbes: a.topologyProbeBundle.GetStatus(),
		FlowProbes:     a.flowProbeBundle.EnabledProbes(),
	}
}

// Start the agent services
func (a *Agent) Start() {
	a.pod.Start()
	a.topologyProbeBundle.Start()
	a.flowProbeBundle.Start()
	a.onDemandPIServer.Start()
	a.onDemandProbeServer.Start()
}

// Stop agent services
func (a *Agent) Stop() {
	a.pod.Stop()
	a.onDemandPIServer.Stop()
	a.onDemandProbeServer.Stop()
	a.flowProbeBundle.Stop()
	a.topologyProbeBundle.Stop()
	a.flowClientPool.Close()

	if tr, ok := http.DefaultTransport.(interface {
		CloseIdleConnections()
	}); ok {
		tr.CloseIdleConnections()
	}

	a.tidMapper.Stop()
}

// NewAgent instantiates a new Agent aiming to launch probes (topology and flow)
func NewAgent() (*Agent, error) {
	if uid := os.Geteuid(); uid != 0 {
		logging.GetLogger().Warning("Agent needs root permissions for some feature like capture, network namespace introspection, some feature might not work as expected")
	}

	backend, err := graph.NewMemoryBackend()
	if err != nil {
		return nil, err
	}

	hostID := config.GetString("host_id")
	origin := graph.Origin(hostID, config.AgentService)
	g := graph.NewGraph(hostID, backend, origin)

	apiAuthBackendName := config.GetString("agent.auth.api.backend")
	apiAuthBackend, err := config.NewAuthenticationBackendByName(apiAuthBackendName)
	if err != nil {
		return nil, err
	}
	clusterAuthOptions := &shttp.AuthenticationOpts{
		Username: config.GetString("agent.auth.cluster.username"),
		Password: config.GetString("agent.auth.cluster.password"),
		Cookie:   config.GetStringMapString("http.cookie"),
	}

	analyzers, err := config.GetAnalyzerServiceAddresses()
	if err != nil {
		return nil, fmt.Errorf("Unable to get the analyzers list: %s", err)
	}

	if len(analyzers) == 0 {
		logging.GetLogger().Info("Agent is running in standalone mode")
	}

	var tlsConfig *tls.Config
	if config.IsTLSEnabled() {
		tlsConfig, err = config.GetTLSServerConfig(true)
		if err != nil {
			return nil, err
		}
	}

	wsClientOpts, err := config.NewWSClientOpts(clusterAuthOptions)
	if err != nil {
		return nil, err
	}

	agent := &Agent{}

	opts := pod.Opts{
		Hubs:                analyzers,
		WebsocketOpts:       config.NewWSServerOpts(),
		WebsocketClientOpts: *wsClientOpts,
		APIValidator:        validator.Validator,
		GraphValidator:      topology.SchemaValidator,
		TLSConfig:           tlsConfig,
		APIAuthBackend:      apiAuthBackend,
		TopologyMarshallers: api.TopologyMarshallers,
		StatusReporter:      agent,
	}

	listenAddr := config.GetString("agent.listen")
	pod, err := pod.NewPod(hostID, config.AgentService, listenAddr, "/ws/agent/topology", g, opts)
	if err != nil {
		return nil, err
	}
	agent.pod = pod

	agent.tidMapper = topology.NewTIDMapper(g)
	agent.tidMapper.Start()

	rootNode, err := createRootNode(g)
	if err != nil {
		return nil, err
	}

	agent.topologyProbeBundle, err = NewTopologyProbeBundle(g, rootNode)
	if err != nil {
		return nil, err
	}

	return agent, nil
}
