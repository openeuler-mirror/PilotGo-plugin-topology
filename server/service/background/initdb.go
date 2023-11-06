package service

import (
	"os"
	"time"

	"github.com/pkg/errors"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
)

func InitDB() {
	var graphperiod int64
	var runningAgents int

	graphperiod = conf.Global_config.Topo.Period
	switch conf.Global_config.Topo.GraphDB {
	case "neo4j":
		dao.Neo4j = dao.CreateNeo4j(conf.Global_config.Neo4j.Addr, conf.Global_config.Neo4j.Username, conf.Global_config.Neo4j.Password, conf.Global_config.Neo4j.DB)
		driver, err := dao.Neo4j.Create_driver()
		if err != nil {
			err := errors.Errorf("create neo4j driver failed: %s **fatal**2", err.Error()) // err top
			agentmanager.Topo.ErrCh <- err
			agentmanager.Topo.Errmu.Lock()
			agentmanager.Topo.ErrCond.Wait()
			agentmanager.Topo.Errmu.Unlock()
			close(agentmanager.Topo.ErrCh)
			os.Exit(1)
		}
		dao.Neo4j.Driver = driver
		
		go func(interval int64) {
			if conf.Config().Topo.Use {
				for true {
					// if runningAgents = agentmanager.Topo.GetRunningAgentNumber(); runningAgents <= 0 {
					// 	err := errors.New("no running agent **warn**1")
					// 	agentmanager.Topo.ErrCh <- err

					// 	time.Sleep(5 * time.Second)
					// 	continue
					// }

					unixtime_now := time.Now().Unix()
					PeriodProcessNeo4j(unixtime_now, runningAgents)
					time.Sleep(time.Duration(interval) * time.Second)

					// break
				}
			}
		}(graphperiod)
	case "otherDB":

	default:
		err := errors.Errorf("unknown database in config_server.yaml: %s **fatal**4", conf.Global_config.Topo.GraphDB) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.Errmu.Lock()
		agentmanager.Topo.ErrCond.Wait()
		agentmanager.Topo.Errmu.Unlock()
		close(agentmanager.Topo.ErrCh)
		os.Exit(1)
	}
}
