package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/graffiti/service"
	"github.com/skydive-project/skydive/graffiti/logging"
	"go.etcd.io/etcd/client/pkg/v3/types"
	client "go.etcd.io/etcd/client/v2"
	"go.etcd.io/etcd/pkg/v3/osutil"
	"go.etcd.io/etcd/server/v3/embed"
)

const (
	startTimeout = 10 * time.Second
)

// EmbeddedServer provides a single node etcd server.
type EmbeddedServer struct {
	Port   int
	config *embed.Config
	etcd   *embed.Etcd
	logger logging.Logger
}

// EmbeddedServerOpts describes the options for an embedded etcd server
type EmbeddedServerOpts struct {
	Name         string
	Listen       string
	Peers        map[string]string
	DataDir      string
	MaxWalFiles  uint
	MaxSnapFiles uint
	Debug        bool
	Logger       logging.Logger
}

// NewEmbeddedServer creates a new embedded ETCD server
func NewEmbeddedServer(opts EmbeddedServerOpts) (*EmbeddedServer, error) {
	if opts.Logger == nil {
		opts.Logger = logging.GetLogger()
	}

	sa, err := service.AddressFromString(opts.Listen)
	if err != nil {
		return nil, err
	}

	cfg := embed.NewConfig()
	cfg.Name = opts.Name
	cfg.Dir = opts.DataDir
	cfg.EnableV2 = true
	cfg.ClusterState = embed.ClusterStateFlagNew
	cfg.MaxWalFiles = opts.MaxWalFiles
	cfg.MaxSnapFiles = opts.MaxSnapFiles

	var listenClientURLs types.URLs
	var listenPeerURLs types.URLs
	if sa.Addr == "0.0.0.0" || sa.Addr == "::" {
		if listenClientURLs, err = interfaceURLs(sa.Port); err != nil {
			return nil, err
		}
		if listenPeerURLs, err = interfaceURLs(sa.Port + 1); err != nil {
			return nil, err
		}
	} else {
		listenClientURLs, _ = types.NewURLs([]string{fmt.Sprintf("http://%s:%d", sa.Addr, sa.Port)})
		listenPeerURLs, _ = types.NewURLs([]string{fmt.Sprintf("http://%s:%d", sa.Addr, sa.Port+1)})
	}

	cfg.ListenClientUrls = listenClientURLs
	cfg.ListenPeerUrls = listenPeerURLs
	cfg.AdvertiseClientUrls = listenClientURLs // This probably won't work with proxy feature

	var advertisePeerUrls types.URLs
	if len(opts.Peers) != 0 {
		initialPeers, err := types.NewURLsMapFromStringMap(opts.Peers, ",")
		if err != nil {
			return nil, err
		}

		if advertisePeerUrls = initialPeers[opts.Name]; advertisePeerUrls == nil {
			return nil, fmt.Errorf("Unable to find Etcd name entry in the peers list: %s", opts.Name)
		}
		cfg.InitialCluster = initialPeers.String()
	}

	if advertisePeerUrls == nil {
		advertisePeerUrls, _ = types.NewURLs([]string{fmt.Sprintf("http://localhost:%d", sa.Port+1)})
		cfg.InitialCluster = types.URLsMap{opts.Name: advertisePeerUrls}.String()
	}

	cfg.AdvertisePeerUrls = advertisePeerUrls

	return &EmbeddedServer{
		Port:   sa.Port,
		config: cfg,
		logger: opts.Logger,
	}, nil
}

// Start etcd server
func (se *EmbeddedServer) Start() error {
	cfg := se.config
	etcd, err := embed.StartEtcd(cfg)
	if err != nil {
		return err
	}
	se.etcd = etcd

	osutil.RegisterInterruptHandler(etcd.Close)

	select {
	case <-etcd.Server.ReadyNotify():
		se.logger.Infof("Server is ready!")
	case <-time.After(60 * time.Second):
		etcd.Server.Stop() // trigger a shutdown
		se.logger.Errorf("Server took too long to start!")
	}

	// Wait for etcd server to be ready
	t := time.Now().Add(startTimeout)

	clientConfig := client.Config{
		Endpoints:               types.URLs(cfg.ListenClientUrls).StringSlice(),
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	etcdClient, err := client.New(clientConfig)
	if err != nil {
		return err
	}
	kapi := client.NewKeysAPI(etcdClient)

	for {
		if time.Now().After(t) {
			return errors.New("Failed to start etcd")
		}
		if _, err := kapi.Set(context.Background(), "/"+se.config.Name, "", nil); err == nil {
			se.logger.Infof("Successfully started etcd")
			break
		}
		time.Sleep(time.Second)
	}

	return nil
}
