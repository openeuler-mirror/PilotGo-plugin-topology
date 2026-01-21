package agent

import (
	"fmt"

	"github.com/skydive-project/skydive/plugin"
	"github.com/skydive-project/skydive/probe"
	"github.com/skydive-project/skydive/topology"
	"github.com/skydive-project/skydive/topology/probes/bess"
	"github.com/skydive-project/skydive/topology/probes/blockdev"
	"github.com/skydive-project/skydive/topology/probes/hardware"
	"github.com/skydive-project/skydive/topology/probes/neutron"
	"github.com/skydive-project/skydive/topology/probes/ovsdb"
)

func registerStaticProbes() {
	blockdev.Register()
	netlink.Register()
	topology.RegisterContainer()
	lldp.Register()
	lxd.Register()
	neutron.Register()
	opencontrail.Register()
	ovsdb.Register()
	libvirt.Register()
	ovn.Register()
	hardware.Register()
}

// NewTopologyProbe creates a new topology probe
func NewTopologyProbe(name string, ctx tp.Context, bundle *probe.Bundle) (probe.Handler, error) {
	switch name {
	case "blockdev":
		return blockdev.NewProbe(ctx, bundle)
	case "netlink":
		return netlink.NewProbe(ctx, bundle)
	case "netns":
		return netns.NewProbe(ctx, bundle)
	case "ovsdb":
		return ovsdb.NewProbe(ctx, bundle)
	case "lxd":
		return lxd.NewProbe(ctx, bundle)
	case "docker":
		return docker.NewProbe(ctx, bundle)
	case "lldp":
		return lldp.NewProbe(ctx, bundle)
	case "neutron":
		return neutron.NewProbe(ctx, bundle)
	case "opencontrail":
		return opencontrail.NewProbe(ctx, bundle)
	case "socketinfo":
		return socketinfo.NewProbe(ctx, bundle)
	case "libvirt":
		return libvirt.NewProbe(ctx, bundle)
	case "runc":
		return runc.NewProbe(ctx, bundle)
	case "vpp":
		return vpp.NewProbe(ctx, bundle)
	case "bess":
		return bess.NewProbe(ctx, bundle)
	case "hardware":
		return hardware.NewProbe(ctx, bundle)
	default:
		return nil, fmt.Errorf("unsupported probe %s", name)
	}
}

func loadPluginProbes(ctx tp.Context, bundle *probe.Bundle) error {
	plugins, err := plugin.LoadTopologyPlugins()
	if err != nil {
		return err
	}

	for _, p := range plugins {
		p.Register()
	}

	// then register agent plugins
	for _, p := range plugins {
		if p.AgentCtor != nil {
			handler, err := p.AgentCtor(ctx, bundle)
			if err != nil {
				return fmt.Errorf("Failed to instantiate plugin %s: %s", p.Name, err)
			}

			bundle.AddHandler(p.Name, handler)
		}
	}

	return nil
}
