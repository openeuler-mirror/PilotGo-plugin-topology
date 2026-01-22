package server

import (
	"net/http"
	"time"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/rbac"
	auth "github.com/abbot/go-http-auth"
	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/flow"
	"github.com/skydive-project/skydive/flow/storage"
	shttp "github.com/skydive-project/skydive/graffiti/http"
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

func (p *PcapAPI) injectPcap(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	updateEvery := time.Duration(config.GetInt("flow.update")) * time.Second
	expireAfter := time.Duration(config.GetInt("flow.expire")) * time.Second

	if !rbac.Enforce(r.Username, "pcap", "write") {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	flowtable := flow.NewTable(updateEvery, expireAfter, p, flow.UUIDs{}, flow.TableOpts{})
	packetSeqChan, _, _ := flowtable.Start(nil)

	feeder, err := flow.NewPcapTableFeeder(r.Body, packetSeqChan, false, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	feeder.Start()
	feeder.Wait()

	// stop/flush flowtable
	flowtable.Stop()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusAccepted)
}

func (p *PcapAPI) registerEndpoints(r *shttp.Server, authBackend shttp.AuthenticationBackend) {
	routes := []shttp.Route{
		{
			Name:        "PCAP",
			Method:      "POST",
			Path:        "/api/pcap",
			HandlerFunc: p.injectPcap,
		},
	}

	r.RegisterRoutes(routes, authBackend)
}

// RegisterPcapAPI registers a new pcap injector API
func RegisterPcapAPI(r *shttp.Server, store storage.Storage, authBackend shttp.AuthenticationBackend) {
	// swagger:operation POST /pcap injectPCAP
	//
	// Inject PCAP
	//
	// ---
	// summary: Inject PCAP
	//
	// tags:
	// - PCAP
	//
	// consumes:
	// - application/octet-stream
	//
	// schemes:
	// - http
	// - https
	//
	// parameters:
	//   - in: body
	//     name: status
	//     required: true
	//     schema:
	//       type: string
	//       format: binary
	//
	// responses:
	//   202:
	//     description: request accepted
	//   400:
	//     description: invalid PCAP

	p := &PcapAPI{
		storage: store,
	}

	p.registerEndpoints(r, authBackend)
}
