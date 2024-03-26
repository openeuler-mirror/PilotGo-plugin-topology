package service

import (
	"time"

	"github.com/pkg/errors"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func InitDB() {
	initGraphDB()

	initRedis()

	initMysql()

	go ClearGraphData(conf.Config().Topo.Retention)
}

// 初始化图数据库
func initGraphDB() {
	switch conf.Global_config.Topo.GraphDB {
	case "neo4j":
		dao.Global_Neo4j = dao.Neo4jInit(conf.Global_config.Neo4j.Addr, conf.Global_config.Neo4j.Username, conf.Global_config.Neo4j.Password, conf.Global_config.Neo4j.DB)
		dao.Global_GraphDB = dao.Global_Neo4j
	case "otherDB":

	default:
		err := errors.Errorf("unknown database in topo_server.yaml: %s **errstackfatal**4", conf.Global_config.Topo.GraphDB) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, true)
	}

	if dao.Global_GraphDB != nil {
		logger.Debug("graph database initialization successful")
	} else {
		logger.Error("graph database initialization failed")
	}
}

// 初始化redis
func initRedis() {
	dao.Global_redis = dao.RedisInit(conf.Config().Redis.Addr, conf.Config().Redis.Password, conf.Config().Redis.DB, conf.Config().Redis.DialTimeout)
	if dao.Global_redis != nil {
		logger.Debug("redis database initialization successful")
	} else {
		logger.Error("redis database initialization failed")
	}
}

func initMysql() {
	dao.Global_mysql = dao.MysqldbInit(conf.Config().Mysql)
	if dao.Global_mysql != nil {
		logger.Debug("mysql database initialization successful")
	} else {
		logger.Error("mysql database initialization failed")
	}
}

func ClearGraphData(retention int64) {
	dao.Global_GraphDB.ClearExpiredData(retention)

	for {
		current := time.Now()
		clear, err := time.Parse("15:04:05", conf.Config().Topo.Cleartime)
		if err != nil {
			logger.Error("ClearGraphData time parse error: %s, %s", err.Error(), conf.Config().Topo.Cleartime)
		}

		next := time.Date(current.Year(), current.Month(), current.Day()+1, clear.Hour(), clear.Minute(), clear.Second(), 0, current.Location())
		if next.Before(current) {
			next = next.Add(24 * time.Hour)
		}

		timer := time.NewTimer(next.Sub(current))

		<-timer.C

		dao.Global_GraphDB.ClearExpiredData(retention)
	}
}
