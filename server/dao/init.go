package dao

import (
	"os"
	"time"

	"github.com/pkg/errors"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
)

func InitDB() {
	var unixtime_now int64
	var period int64
	var runningAgents int
	period = conf.Global_config.Topo.Period

	switch conf.Global_config.Topo.Database {
	case "neo4j":
		go func(pt int64) {
			for true {
				// if runningAgents = agentmanager.Topo.GetRunningAgentNumber(); runningAgents <= 0 {
				// 	err := errors.New("no running agent **warn**1")
				// 	agentmanager.Topo.ErrCh <- err

				// 	time.Sleep(5 * time.Second)
				// 	continue
				// }

				unixtime_now = time.Now().Unix()
				Neo4j = CreateNeo4j(conf.Global_config.Neo4j.Addr, conf.Global_config.Neo4j.Username, conf.Global_config.Neo4j.Password, conf.Global_config.Neo4j.DB)
				PeriodProcessNeo4j(unixtime_now, runningAgents)
				time.Sleep(time.Duration(period) * time.Second)

				// break
			}
		}(period)
	case "otherDB":

	default:
		err := errors.Errorf("unknown database in config_server.yaml: %s **fatal**4", conf.Global_config.Topo.Database) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.Errmu.Lock()
		agentmanager.Topo.ErrCond.Wait()
		agentmanager.Topo.Errmu.Unlock()
		close(agentmanager.Topo.ErrCh)
		os.Exit(1)
	}
}
