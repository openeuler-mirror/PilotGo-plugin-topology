package agent

import (
	"errors"
	"os"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"github.com/skydive-project/skydive/config"
)

// createRootNode creates a graph.Node based on the host properties and aims to have an unique ID
func createRootNode(g *graph.Graph) (*graph.Node, error) {
	hostID := config.GetString("host_id")
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	m := graph.Metadata{"Name": hostID, "Type": "host", "Hostname": hostname}

	// Fill the metadata from the configuration file
	if configMetadata := config.Get("agent.metadata"); configMetadata != nil {
		configMetadata, ok := graph.NormalizeValue(configMetadata).(map[string]interface{})
		if !ok {
			return nil, errors.New("agent.metadata has wrong format")
		}
		for k, v := range configMetadata {
			m[k] = v
		}
	}

	hostNode, err := g.NewNode(graph.GenID(), m)
	if err != nil {
		return nil, err
	}

	parseMetadataConfigFiles(g, hostNode)

	return hostNode, nil
}

type metadataConfigFile struct {
	Path     string
	Type     string
	Name     string
	Selector string
}
