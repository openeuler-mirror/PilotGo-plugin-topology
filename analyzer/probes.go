package analyzer

import (
	fp "github.com/skydive-project/skydive/flow/probes"
	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/packetinjector"
	"github.com/skydive-project/skydive/plugin"
	"github.com/skydive-project/skydive/sflow"
	"github.com/skydive-project/skydive/topology"
	"github.com/skydive-project/skydive/topology/probes/blockdev"
	"github.com/skydive-project/skydive/topology/probes/libvirt"
	"github.com/skydive-project/skydive/topology/probes/lldp"
	"github.com/skydive-project/skydive/topology/probes/lxd"
	"github.com/skydive-project/skydive/topology/probes/netlink"
	"github.com/skydive-project/skydive/topology/probes/neutron"
	"github.com/skydive-project/skydive/topology/probes/opencontrail"
	"github.com/skydive-project/skydive/topology/probes/ovn"
	"github.com/skydive-project/skydive/topology/probes/ovsdb"
)

func registerStaticProbes() {
	netlink.Register()
	blockdev.Register()
	topology.RegisterContainer()
	lldp.Register()
	lxd.Register()
	neutron.Register()
	opencontrail.Register()
	ovsdb.Register()
	libvirt.Register()
	ovn.Register()
}

func registerPluginProbes() error {
	plugins, err := plugin.LoadTopologyPlugins()
	if err != nil {
		return err
	}

	for _, p := range plugins {
		p.Register()
	}

	return nil
}

// RegisterProbes register graph metadata decoders
func registerProbes() error {
	registerStaticProbes()

	if err := registerPluginProbes(); err != nil {
		return err
	}

	graph.NodeMetadataDecoders["Captures"] = fp.CapturesMetadataDecoder
	graph.NodeMetadataDecoders["PacketInjections"] = packetinjector.InjectionsMetadataDecoder

	// TODO move it when flow probe plugin will be introduced
	graph.NodeMetadataDecoders["SFlow"] = sflow.SFMetadataDecoder

	return nil
}
