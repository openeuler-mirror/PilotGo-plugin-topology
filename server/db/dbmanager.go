package db

import (
	"time"

	"github.com/pkg/errors"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/graphmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/influxmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/redismanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func InitDB() {
	if conf.Global_Config.Topo.GraphDB != "" {
		initGraphDB()
		ClearGraphData(conf.Global_Config.Neo4j.Retention)
	} else {
		err := errors.New("do not save graph data **warn**0")
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
	}

	initRedis()

	initMysql()

	// initInflux()
}

// 初始化图数据库
func initGraphDB() {
	global.Global_graph_database = conf.Global_Config.Topo.GraphDB

	switch conf.Global_Config.Topo.GraphDB {
	case "neo4j":
		graphmanager.Global_Neo4j = graphmanager.Neo4jInit(conf.Global_Config.Neo4j.Addr, conf.Global_Config.Neo4j.Username, conf.Global_Config.Neo4j.Password, conf.Global_Config.Neo4j.DB)
		graphmanager.Global_GraphDB = graphmanager.Global_Neo4j
	case "otherDB":

	default:
		err := errors.Errorf("unknown database in topo_server.yaml: %s **errstackfatal**4", conf.Global_Config.Topo.GraphDB) // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
	}

	if graphmanager.Global_GraphDB != nil {
		logger.Debug("graph database initialization successful")
	} else {
		logger.Error("graph database initialization failed")
	}
}

// 初始化redis
func initRedis() {
	redismanager.Global_Redis = redismanager.RedisInit(conf.Global_Config.Redis.Addr, conf.Global_Config.Redis.Password, conf.Global_Config.Redis.DB, conf.Global_Config.Redis.DialTimeout)
	if redismanager.Global_Redis != nil {
		logger.Debug("redis database initialization successful")
	} else {
		logger.Error("redis database initialization failed")
	}
}

func initMysql() {
	mysqlmanager.Global_Mysql = mysqlmanager.MysqldbInit(conf.Global_Config.Mysql)
	if mysqlmanager.Global_Mysql != nil {
		logger.Debug("mysql database initialization successful")
	} else {
		logger.Error("mysql database initialization failed")
	}
}

func initInflux() {
	influxmanager.Global_Influx = influxmanager.InfluxdbInit(conf.Global_Config.Influx)
	if influxmanager.Global_Influx != nil {
		logger.Debug("influx database initialization successful")
	} else {
		logger.Error("influx database initialization failed")
	}
}

func ClearGraphData(retention int64) {
	if graphmanager.Global_GraphDB == nil {
		err := errors.New("global_graphdb is nil **errstackfatal**0")
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	graphmanager.Global_GraphDB.ClearExpiredData(retention)

	global.Global_wg.Add(1)
	go func() {
		defer global.Global_wg.Done()
		for {
			select {
			case <-global.Global_cancelCtx.Done():
				return
			default:
				current := time.Now()
				clear, err := time.Parse("15:04:05", conf.Global_Config.Neo4j.Cleartime)
				if err != nil {
					logger.Error("ClearGraphData time parse error: %s, %s", err.Error(), conf.Global_Config.Neo4j.Cleartime)
				}

				next := time.Date(current.Year(), current.Month(), current.Day()+1, clear.Hour(), clear.Minute(), clear.Second(), 0, current.Location())
				if next.Before(current) {
					next = next.Add(24 * time.Hour)
				}

				timer := time.NewTimer(next.Sub(current))

				<-timer.C

				graphmanager.Global_GraphDB.ClearExpiredData(retention)
			}

		}
	}()

}
