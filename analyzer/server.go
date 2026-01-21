package analyzer

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	api "github.com/skydive-project/skydive/api/server"
	"github.com/skydive-project/skydive/api/types"
	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/flow/server"
	"github.com/skydive-project/skydive/flow/storage"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	gapi "github.com/skydive-project/skydive/graffiti/api/server"
	gtypes "github.com/skydive-project/skydive/graffiti/api/types"
	etcdclient "github.com/skydive-project/skydive/graffiti/etcd/client"
	"github.com/skydive-project/skydive/graffiti/graph"
	shttp "github.com/skydive-project/skydive/graffiti/http"
	"github.com/skydive-project/skydive/graffiti/hub"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/graffiti/ondemand/client"
	ws "github.com/skydive-project/skydive/graffiti/websocket"
	"github.com/skydive-project/skydive/probe"
	"github.com/skydive-project/skydive/statics"
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

// createStartupCapture creates capture based on preconfigured selected SubGraph
func (s *Server) createStartupCapture() error {
	apiHandler := s.hub.APIServer().GetHandler("capture").(*api.CaptureAPIHandler)

	gremlin := config.GetString("analyzer.startup.capture_gremlin")
	if gremlin == "" {
		return nil
	}

	bpf := config.GetString("analyzer.startup.capture_bpf")
	captureType := config.GetString("analyzer.startup.capture_type")
	logging.GetLogger().Infof("Invoke capturing of type '%s' from startup with gremlin: %s and BPF: %s", captureType, gremlin, bpf)
	capture := types.NewCapture(gremlin, bpf)
	capture.Type = captureType
	if err := apiHandler.Create(capture, nil); err != nil && !errors.Is(err, rest.ErrDuplicatedResource) {
		return err
	}
	return nil
}

func (s *Server) loadStaticWorkflows() error {
	assets, err := statics.AssetDir(workflowAssetDir)
	if err == nil {
		for _, asset := range assets {
			yml, err := statics.Asset(workflowAssetDir + "/" + asset)
			if err != nil {
				return err
			}

			var workflow gtypes.Workflow
			if err := yaml.Unmarshal([]byte(yml), &workflow); err != nil {
				return errors.Wrapf(err, "failed to load workflow %s", asset)
			}

			gapi.StaticWorkflows = append(gapi.StaticWorkflows, &workflow)
		}
	}

	return nil
}

// Start the analyzer server
func (s *Server) Start() error {
	if err := s.hub.Start(); err != nil {
		return err
	}

	s.etcdClient.Start()

	if s.flowStorage != nil {
		s.flowStorage.Start()
	}

	if err := s.probeBundle.Start(); err != nil {
		return err
	}

	s.onDemandClient.Start()
	s.piClient.Start()
	s.topologyManager.Start()
	s.flowServer.Start()

	if err := s.createStartupCapture(); err != nil {
		return err
	}

	return nil
}

// Stop the analyzer server
func (s *Server) Stop() {
	s.hub.Stop()
	s.flowServer.Stop()
	s.probeBundle.Stop()
	s.onDemandClient.Stop()
	s.piClient.Stop()
	s.topologyManager.Stop()
	s.etcdClient.Stop()

	if s.flowStorage != nil {
		s.flowStorage.Stop()
	}

	if tr, ok := http.DefaultTransport.(interface {
		CloseIdleConnections()
	}); ok {
		tr.CloseIdleConnections()
	}
}

// NewServerFromConfig creates a new empty server
func NewServerFromConfig() (*Server, error) {
	host := config.GetString("host_id")

	etcdClientOpts := etcdclient.Opts{
		Servers: config.GetEtcdServerAddrs(),
		Timeout: time.Duration(config.GetInt("etcd.client_timeout")) * time.Second,
	}

	etcdClient, err := etcdclient.NewClient(host, etcdClientOpts)
	if err != nil {
		return nil, err
	}

	if err := config.InitRBAC(etcdClient.KeysAPI); err != nil {
		return nil, err
	}

	var tlsConfig *tls.Config
	if config.IsTLSEnabled() {
		tlsConfig, err = config.GetTLSServerConfig(true)
		if err != nil {
			return nil, err
		}
	}
	s := &Server{
		probeBundle:  probeBundle,
		etcdClient:   etcdClient,
		graphStorage: graphStorage,
	}
	return s, nil
}

// ClusterAuthenticationOpts returns auth info to connect to an analyzer
// from the configuration
func ClusterAuthenticationOpts() *shttp.AuthenticationOpts {
	return &shttp.AuthenticationOpts{
		Username: config.GetString("analyzer.auth.cluster.username"),
		Password: config.GetString("analyzer.auth.cluster.password"),
		Cookie:   config.GetStringMapString("http.cookie"),
	}
}
