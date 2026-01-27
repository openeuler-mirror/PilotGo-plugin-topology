package config

import (
	"fmt"

	"github.com/skydive-project/skydive/config"
)

const (
	defaultConfigurationFile = "/etc/skydive/skydive.yml"
)

// LoadConfiguration from a configuration file
// If no configuration file are given, try to load the default configuration file /etc/skydive/skydive.yml
func LoadConfiguration(cfgBackend string, cfgFiles []string) error {
	if len(cfgFiles) == 0 {
		config.InitConfig(cfgBackend, []string{defaultConfigurationFile})
		if err := config.InitLogging(); err != nil {
			return fmt.Errorf("Failed to initialize logging system: %s", err.Error())
		}
		return nil
	}

	if err := config.InitConfig(cfgBackend, cfgFiles); err != nil {
		return fmt.Errorf("Failed to initialize config: %s", err.Error())
	}

	if err := config.InitLogging(); err != nil {
		return fmt.Errorf("Failed to initialize logging system: %s", err.Error())
	}

	return nil
}
