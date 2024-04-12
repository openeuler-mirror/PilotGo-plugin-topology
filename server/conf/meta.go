package conf

import "time"

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

type MysqlConf struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string `yaml:"DB"`
}
