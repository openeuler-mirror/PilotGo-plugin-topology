package service

import (
	"strconv"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
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

	if pluginclient.GlobalClient == nil {
		err := errors.New("globalclient is nil **errstackfatal**2") // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
		return nil, nil, nil, err
	}

	machine_uuids, err := pluginclient.GlobalClient.BatchUUIDList(strconv.Itoa(int(tc.BatchId)))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "**2")
	}

	// ctxv := context.WithValue(agentmanager.Topo.Tctx, "custom_name", "pilotgo-topo")

	if agentmanager.GlobalAgentManager == nil {
		err := errors.New("globalagentmanager is nil **errstackfatal**0") // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
		return nil, nil, nil, err
	}

	agentmanager.GlobalAgentManager.UpdateMachineList()
	dao.Global_redis.ActiveHeartbeatDetection(machine_uuids)
	running_agent_num := dao.Global_redis.UpdateTopoRunningAgentList(machine_uuids, true)
	if running_agent_num == 0 {
		return nil, nil, nil, errors.Errorf("no running agent for custom id %d **errstack**2", tc.ID)
	} else if running_agent_num == -1 {
		return nil, nil, nil, errors.New("redis client not init **errstack**1")
	}

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

func UpdateCustomTopoService(tc *meta.Topo_configuration, tcdb_id_old uint) (int, error) {
	tcdb, err := dao.Global_mysql.TopoConfigurationToDB(tc)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	tcdb_old, err := dao.Global_mysql.QuerySingleTopoConfiguration(tcdb_id_old)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	tcdb_old.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	tcdb_old.BatchId = tcdb.BatchId
	tcdb_old.TagRules = tcdb.TagRules
	tcdb_old.NodeRules = tcdb.NodeRules
	tcdb_old.Name = tcdb.Name
	tcdb_old.Description = tcdb.Description
	tcdb_old.Version = tcdb.Version
	tcdb_old.Preserve = tcdb.Preserve

	tcdb_old_id, err := dao.Global_mysql.AddTopoConfiguration(tcdb_old)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	return tcdb_old_id, nil
}

func DeleteCustomTopoService(ids []uint) error {
	for _, tcid := range ids {
		if err := dao.Global_mysql.DeleteTopoConfiguration(tcid); err != nil {
			return errors.Wrap(err, "**errstack**2")
		}
	}

	return nil
}
