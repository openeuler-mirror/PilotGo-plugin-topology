package service

import (
	"os"

	"github.com/pkg/errors"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
)

func InitDB() {
	switch conf.Global_config.Topo.GraphDB {
	case "neo4j":
		dao.Global_Neo4j = dao.Neo4jInit(conf.Global_config.Neo4j.Addr, conf.Global_config.Neo4j.Username, conf.Global_config.Neo4j.Password, conf.Global_config.Neo4j.DB)
		dao.Global_GraphDB = dao.Global_Neo4j
	case "otherDB":

	default:
		err := errors.Errorf("unknown database in topo_server.yaml: %s **fatal**4", conf.Global_config.Topo.GraphDB) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.Errmu.Lock()
		agentmanager.Topo.ErrCond.Wait()
		agentmanager.Topo.Errmu.Unlock()
		close(agentmanager.Topo.ErrCh)
		os.Exit(1)
	}

	dao.Global_redis = dao.RedisInit(conf.Config().Redis.Addr, conf.Config().Redis.Password, conf.Config().Redis.DB, conf.Config().Redis.DialTimeout)

}
