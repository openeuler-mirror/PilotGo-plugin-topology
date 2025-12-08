package client

import (
	"sync"
	"time"

	"github.com/safchain/insanelock"
	etcd "go.etcd.io/etcd/client/v2"
	"golang.org/x/net/context"

	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/graffiti/service"
)

const (
	timeout = time.Second * 30
)

// MasterElector describes an ETCD master elector
type MasterElector struct {
	insanelock.RWMutex
	EtcdKeyAPI etcd.KeysAPI
	HolderID   string
	path       string
	listeners  []MasterElectionListener
	cancel     context.CancelFunc
	master     bool
	state      service.State
	wg         sync.WaitGroup
	logger     logging.Logger
}

// TTL time to live
func (le *MasterElector) TTL() time.Duration {
	return timeout
}

func (le *MasterElector) holdLock(quit chan bool) {
	defer close(quit)

	tick := time.NewTicker(timeout / 2)
	defer tick.Stop()

	setOptions := &etcd.SetOptions{
		TTL:       timeout,
		PrevExist: etcd.PrevExist,
		PrevValue: le.HolderID,
	}

	ch := tick.C

	for {
		select {
		case <-ch:
			if _, err := le.EtcdKeyAPI.Set(context.Background(), le.path, le.HolderID, setOptions); err != nil {
				return
			}
		case <-quit:
			return
		}
	}
}
