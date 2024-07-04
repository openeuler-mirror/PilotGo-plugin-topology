package conf

import "time"

type TopoConf struct {
	Https_enabled      bool   `yaml:"https_enabled"`
	Public_certificate string `yaml:"cert_file"`
	Private_key        string `yaml:"key_file"`
	Addr               string `yaml:"addr"`
	Agent_port         string `yaml:"agent_port"`
	GraphDB            string `yaml:"graphDB"`
	Period             int64  `yaml:"period"`
	Retention          int64  `yaml:"retention"`
	Cleartime          string `yaml:"cleartime"`
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
	UseTLS      bool          `yaml:"use_tls"`
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

type InfluxConf struct {
	Addr   string `yaml:"addr"`
	Token  string `yaml:"token"`
	Org    string `yaml:"org"`
	Bucket string `yaml:"bucket"`
}
