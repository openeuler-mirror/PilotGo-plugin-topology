package conf

import (
	"path"
	"time"

	"gitee.com/openeuler/PilotGo/sdk/logger"
)

type TopoConf struct {
	Server_addr string `yaml:"server_addr"`
	Agent_port  string `yaml:"agent_port"`
	GraphDB     string `yaml:"graphDB"`
	Period      int64  `yaml:"period"`
	Retention   int64  `yaml:"retention"`
	Cleartime   string `yaml:"cleartime"`
}

type PilotGoConf struct {
	Addr string `yaml:"http_addr"`
}

type ArangodbConf struct {
	Addr     string `yaml:"addr"`
	Database string `yaml:"database"`
}

type Neo4jConf struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string `yaml:"DB"`
}

type PrometheusConf struct {
	Addr string `yaml:"addr"`
}

type RedisConf struct {
	Addr        string        `yaml:"addr"`
	Password    string        `yaml:"password"`
	DB          int           `yaml:"DB"`
	DialTimeout time.Duration `yaml:"dialTimeout"`
}

type ServerConfig struct {
	Topo       *TopoConf
	PilotGo    *PilotGoConf
	Logopts    *logger.LogOpts `yaml:"log"`
	Arangodb   *ArangodbConf
	Neo4j      *Neo4jConf
	Prometheus *PrometheusConf
	Redis      *RedisConf
}

const config_type = "topo_server.yaml"

var Config_dir string

func Config_file() string {
	// _, thisfilepath, _, _ := runtime.Caller(0)
	// dirpath := filepath.Dir(thisfilepath)
	// configfilepath := path.Join(dirpath, "..", "..", "conf", config_type)

	// ttcode:
	configfilepath := path.Join(Config_dir, config_type)

	return configfilepath
}

var Global_config ServerConfig

func Config() *ServerConfig {
	return &Global_config
}
