package agent

import (
	"errors"
	"fmt"
	"os"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/spf13/viper"
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

func parseMetadataConfigFiles(g *graph.Graph, hostNode *graph.Node) error {
	files := config.Get("agent.metadata_config.files")
	if files == nil {
		return nil
	}

	var mcfs []metadataConfigFile
	if err := mapstructure.Decode(files, &mcfs); err != nil {
		return fmt.Errorf("Unable to read agent.metadata_config.files: %s", err)
	}

	updatedHostNode := func(v *viper.Viper, mcf metadataConfigFile) {
		s := v.GetString(mcf.Selector)

		g.Lock()
		g.AddMetadata(hostNode, fmt.Sprintf("Config.%s", mcf.Name), s)
		g.Unlock()
	}

	for _, mcf := range mcfs {
		v := viper.New()
		v.SetConfigFile(mcf.Path)

		typ := mcf.Type
		if typ == "ini" {
			typ = "toml"
		}

		v.SetConfigType(typ)
		v.OnConfigChange(func(in fsnotify.Event) {
			updatedHostNode(v, mcf)
		})
		v.WatchConfig()
		if err := v.ReadInConfig(); err != nil {
			logging.GetLogger().Errorf("Unable to read %s: %s", mcf.Path, err)
			continue
		}
		updatedHostNode(v, mcf)
	}

	return nil
}
