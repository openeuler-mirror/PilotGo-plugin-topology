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
	period = conf.Global_config.Topo.Period

	switch conf.Global_config.Topo.Database {
	case "neo4j":
		go func(pt int64) {
			for true {
				unixtime_now = time.Now().Unix()

				PeriodProcessNeo4j(unixtime_now)

				time.Sleep(time.Duration(period) * time.Second)

				// break
			}
		}(period)
	case "otherDB":

	default:
		err := errors.Errorf("unknown database in config_server.yaml: %s **fatal**4", conf.Global_config.Topo.Database) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.ErrGroup.Add(1)
		agentmanager.Topo.ErrGroup.Wait()
		close(agentmanager.Topo.ErrCh)
		os.Exit(1)
	}
}
