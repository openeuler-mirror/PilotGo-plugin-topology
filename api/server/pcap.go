package server

import (
	"github.com/skydive-project/skydive/flow"
	"github.com/skydive-project/skydive/flow/storage"
	"github.com/skydive-project/skydive/graffiti/logging"
)

// PcapAPI exposes the pcap injector API
type PcapAPI struct {
	storage storage.Storage
}

// SendFlows implements the flow Sender interface
func (p *PcapAPI) SendFlows(flows []*flow.Flow) {
	if p.storage != nil && len(flows) > 0 {
		p.storage.StoreFlows(flows)
		logging.GetLogger().Debugf("%d flows stored", len(flows))
	}
}

// SendStats implements the flow Sender interface
func (p *PcapAPI) SendStats(status flow.Stats) {
}
