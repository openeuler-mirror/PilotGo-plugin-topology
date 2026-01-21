package agent

import (
	"atomgit.com/openeuler/PilotGo-plugin-topology/config"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"github.com/skydive-project/skydive/flow"
	fp "github.com/skydive-project/skydive/flow/probes"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/probe"
)

func NewFlowProbeBundle(tb *probe.Bundle, g *graph.Graph, fta *flow.TableAllocator) *probe.Bundle {
	list := []string{"pcapsocket", "ovssflow", "sflow", "gopacket", "dpdk", "ebpf", "ovsmirror", "ovsnetflow"}
	logging.GetLogger().Infof("Flow probes: %v", list)

	var handler fp.FlowProbeHandler
	var err error

	bundle := probe.NewBundle()
	ctx := fp.Context{
		Logger: logging.GetLogger(),
		Config: config.GetConfig(),
		Graph:  g,
		FTA:    fta,
		TB:     tb,
	}

	for _, t := range list {
		if bundle.GetHandler(t) != nil {
			continue
		}
	}

	return bundle
}
