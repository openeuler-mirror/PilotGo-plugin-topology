/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package conf

import (
	"flag"
	"fmt"
	"os"
	"path"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var Global_Config *ServerConfig

const config_type = "topo_server.yaml"

var config_dir string

type ServerConfig struct {
	Topo          *TopoConf
	Logopts       *logger.LogOpts `yaml:"log"`
	Arangodb      *ArangodbConf
	Neo4j         *Neo4jConf
	Prometheus    *PrometheusConf
	Redis         *RedisConf
	Mysql         *MysqlConf
	Influx        *InfluxConf
}

func ConfigFile() string {
	configfilepath := path.Join(config_dir, config_type)

	return configfilepath
}

func InitConfig() {
	flag.StringVar(&config_dir, "conf", "./", "topo-server configuration directory")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -conf /PATH/TO/TOPO_SERVER.YAML(default: ./) \n", os.Args[0])
	}
	flag.Parse()

	bytes, err := global.FileReadBytes(ConfigFile())
	if err != nil {
		flag.Usage()
		// err = errors.Wrapf(err, "open file failed: %s, %s", ConfigFile(), err.Error())
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	Global_Config = &ServerConfig{}

	err = yaml.Unmarshal(bytes, Global_Config)
	if err != nil {
		err = errors.Errorf("yaml unmarshal failed: %s", err.Error())
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
