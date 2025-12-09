package server

import (
	"time"

	"go.etcd.io/etcd/server/v3/embed"

	"github.com/skydive-project/skydive/graffiti/logging"
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
