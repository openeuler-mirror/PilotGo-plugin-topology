package service

import (
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	back "gitee.com/openeuler/PilotGo-plugin-topology-server/service/background"
	"gitee.com/openeuler/PilotGo/sdk/response"
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

func CustomTopoListService(query *response.PaginationQ) ([]*meta.Topo_configuration, int, error) {
	tcs := make([]*meta.Topo_configuration, 0)

	tcdbs, total, err := dao.Global_mysql.QueryTopoConfigurationList(query)
	if err != nil {
		return nil, 0, errors.Wrap(err, "**2")
	}

	for _, tcdb := range tcdbs {
		tc, err := dao.Global_mysql.DBToTopoConfiguration(tcdb)
		if err != nil {
			return nil, 0, errors.Wrap(err, "**2")
		}

		tcs = append(tcs, tc)
	}

	return tcs, total, nil
}

func CreateCustomTopoService(topoconfig *meta.Topo_configuration) (int, error) {
	tcdb, err := dao.Global_mysql.TopoConfigurationToDB(topoconfig)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	tcdb.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	tcdb.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	tcdb_id, err := dao.Global_mysql.AddTopoConfiguration(tcdb)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	return tcdb_id, nil
}