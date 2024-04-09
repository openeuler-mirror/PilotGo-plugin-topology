package agentmanager

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var Topo *Topoclient

type Topoclient struct {
}

func (t *Topoclient) InitLogger() {
	err := logger.Init(conf.Config().Logopts)
	if err != nil {
		err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
	}
}

func (t *Topoclient) InitConfig() {
	flag.StringVar(&conf.Config_dir, "conf", "/opt/PilotGo/plugin/topology/server", "topo-server configuration directory")
	flag.Parse()

	bytes, err := os.ReadFile(conf.Config_file())
	if err != nil {
		err = errors.Errorf("open file failed: %s, %s", conf.Config_file(), err.Error()) // err top
		fmt.Printf("%+v\n", err)
		os.Exit(-1)
	}

	err = yaml.Unmarshal(bytes, &conf.Global_config)
	if err != nil {
		err = errors.Errorf("yaml unmarshal failed: %s", err.Error()) // err top
		fmt.Printf("%+v\n", err)
		os.Exit(-1)
	}
}

func (t *Topoclient) SignalMonitoring(neo4jclient neo4j.Driver, redisclient redis.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		s := <-c
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			neo4jclient.Close()
			fmt.Println()
			logger.Info("close the connection to neo4j\n")
			redisclient.Close()
			logger.Info("close the connection to redis\n")
			os.Exit(-1)
		default:
			logger.Warn("unknown signal-> %s\n", s.String())
		}
	}
}
