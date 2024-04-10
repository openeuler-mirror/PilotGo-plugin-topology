package conf

import (
	"flag"
	"fmt"
	"os"
	"path"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var Global_Config *ServerConfig

const config_type = "topo_server.yaml"

var config_dir string

type ServerConfig struct {
	Topo       *TopoConf
	PilotGo    *PilotGoConf
	Logopts    *logger.LogOpts `yaml:"log"`
	Arangodb   *ArangodbConf
	Neo4j      *Neo4jConf
	Prometheus *PrometheusConf
	Redis      *RedisConf
	Mysql      *MysqlConf
}

func ConfigFile() string {
	configfilepath := path.Join(config_dir, config_type)

	return configfilepath
}

func InitConfig() {
	flag.StringVar(&config_dir, "conf", "/opt/PilotGo/plugin/topology/server", "topo-server configuration directory")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -conf /PATH/TO/TOPO_SERVER.YAML(default:/opt/PilotGo/plugin/topology/server) \n", os.Args[0])
	}
	flag.Parse()

	bytes, err := utils.FileReadBytes(ConfigFile())
	if err != nil {
		err = errors.Wrapf(err, "open file failed: %s, %s", ConfigFile(), err.Error()) // err top
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(bytes, Global_Config)
	if err != nil {
		err = errors.Errorf("yaml unmarshal failed: %s", err.Error()) // err top
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
