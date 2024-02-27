package service

import (
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	back "gitee.com/openeuler/PilotGo-plugin-topology-server/service/background"
	"github.com/pkg/errors"
)

func RunCustomTopoService(tcid uint) ([]*meta.Node, []*meta.Edge, []map[string]string, error) {
	tcdb, err := dao.Global_mysql.QuerySingleTopoConfiguration(tcid)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "**2")
	}

	tc, err := dao.Global_mysql.DBToTopoConfiguration(tcdb)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "**2")
	}

	// ctxv := context.WithValue(agentmanager.Topo.Tctx, "custom_name", "pilotgo-topo")

	agentmanager.Topo.UpdateMachineList()
	running_agent_num := dao.Global_redis.UpdateTopoRunningAgentList(tc.Machines)
	unixtime_now := time.Now().Unix()
	nodes, edges, combos, err := back.DataProcessWorking(unixtime_now, running_agent_num, dao.Global_GraphDB, tc.TagRules, tc.NodeRules)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "**2")
	}

	return nodes, edges, combos, nil
}

func CustomTopoListService() ([]*meta.Topo_configuration, error) {
	tcs := make([]*meta.Topo_configuration, 0)

	tcdbs, err := dao.Global_mysql.QueryTopoConfigurationList()
	if err != nil {
		return nil, errors.Wrap(err, "**2")
	}

	for _, tcdb := range tcdbs {
		tc, err := dao.Global_mysql.DBToTopoConfiguration(tcdb)
		if err != nil {
			return nil, errors.Wrap(err, "**2")
		}

		tcs = append(tcs, tc)
	}

	return tcs, nil
}