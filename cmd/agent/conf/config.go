package conf

import (
	"flag"
	"fmt"
	"os"
	"path"

	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type TopoConf struct {
	Https_enabled      bool   `yaml:"https_enabled"`
	Public_certificate string `yaml:"cert_file"`
	Private_key        string `yaml:"key_file"`
	Agent_addr         string `yaml:"agent_addr"`
	Agent_port         string `yaml:"agent_port"`
	Server_addr        string `yaml:"server_addr"`
	Datasource         string `yaml:"datasource"`
	Heartbeat          int    `yaml:"heartbeat"`
}

type PilotGoConf struct {
	Addr string `yaml:"addr"`
}

type ServerConfig struct {
	Topo    *TopoConf       `yaml:"topo"`
	PilotGo *PilotGoConf    `yaml:"PilotGo"`
	Logopts *logger.LogOpts `yaml:"log"`
}

const config_type = "topo_agent.yaml"

var Config_dir string

func config_file() string {
	// _, thisfilepath, _, _ := runtime.Caller(0)
	// dirpath := filepath.Dir(thisfilepath)
	// configfilepath := path.Join(dirpath, "..", "..", "conf", config_type)

	// ttcode:
	configfilepath := path.Join(Config_dir, config_type)
	return configfilepath
}

var global_config ServerConfig

func init() {
	flag.StringVar(&Config_dir, "conf", "/opt/PilotGo/plugin/topology/agent", "topo-agent configuration directory")
	flag.Parse()

	err := readConfig(config_file(), &global_config)
	if err != nil {
		err = errors.Wrap(err, "")
		fmt.Printf("%s\n", errors.Cause(err).Error()) // err top
		// errors.EORE(err)
		os.Exit(-1)
	}
}

func Config() *ServerConfig {
	return &global_config
}

func readConfig(file string, config interface{}) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		err = errors.Errorf("ERROR: %s", err.Error())
		return err
	}

	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		err = errors.Errorf("yaml Unmarshal %s failed", string(bytes))
		return err
	}
	return nil
}
