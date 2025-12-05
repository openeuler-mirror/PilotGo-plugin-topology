package client

import (
	"time"

	etcd "go.etcd.io/etcd/client/v2"

	"github.com/skydive-project/skydive/graffiti/logging"
)

// Client defaults
const (
	DefaultTimeout = 5 * time.Second
	DefaultPort    = 12379
	DefaultServer  = "127.0.0.1"
)

// MasterElectionListener describes the multi election mechanism
type MasterElectionListener interface {
	OnStartAsMaster()
	OnStartAsSlave()
	OnSwitchToMaster()
	OnSwitchToSlave()
}

// MasterElection describes the master election mechanism
type MasterElection interface {
	Start()
	StartAndWait()
	Stop()
	IsMaster() bool
	AddEventListener(listener MasterElectionListener)
	TTL() time.Duration
}

// MasterElectionService describes the election service mechanism
type MasterElectionService interface {
	NewElection(key string) MasterElection
}

// Client describes a ETCD configuration client
type Client struct {
	id      string
	client  *etcd.Client
	KeysAPI etcd.KeysAPI
	logger  logging.Logger
}

// Opts describes the options of an etcd client
type Opts struct {
	Servers []string
	Timeout time.Duration
	Logger  logging.Logger
}
