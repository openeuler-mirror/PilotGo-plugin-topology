package service

import (
	"strconv"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/graphmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/redismanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/graph"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	back "gitee.com/openeuler/PilotGo-plugin-topology-server/service/background"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/pkg/errors"
)

func RunCustomTopoService(tcid uint) ([]*graph.Node, []*graph.Edge, []map[string]string, error) {
	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil **errstackfatal**2")
		return nil, nil, nil, err
	}

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil **errstackfatal**0")
		return nil, nil, nil, err
	}

	if redismanager.Global_Redis == nil {
		err := errors.New("global_redis is nil **errstackfatal**1")
		return nil, nil, nil, err
	}

	if mysqlmanager.Global_Mysql == nil {
		err := errors.New("global_mysql is nil **errstackfatal**1")
		return nil, nil, nil, err
	}

	tcdb, err := mysqlmanager.Global_Mysql.QuerySingleTopoConfiguration(tcid)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "**2")
	}

	tc, err := mysqlmanager.Global_Mysql.DBToTopoConfiguration(tcdb)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "**2")
	}

	machine_uuids, err := pluginclient.Global_Client.BatchUUIDList(strconv.Itoa(int(tc.BatchId)))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "**2")
	}

	// ctxv := context.WithValue(agentmanager.Topo.Tctx, "custom_name", "pilotgo-topo")

	agentmanager.Global_AgentManager.UpdateMachineList()

	redismanager.Global_Redis.ActiveHeartbeatDetection(machine_uuids)
	running_agent_num := redismanager.Global_Redis.UpdateTopoRunningAgentList(machine_uuids, true)
	if running_agent_num == 0 {
		return nil, nil, nil, errors.Errorf("no running agent for custom id %d **errstack**2", tc.ID)
	} else if running_agent_num == -1 {
		return nil, nil, nil, errors.New("redis client not init **errstack**1")
	}

	unixtime_now := time.Now().Unix()
	nodes, edges, combos, err := back.DataProcessWorking(unixtime_now, running_agent_num, graphmanager.Global_GraphDB, tc.TagRules, tc.NodeRules)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "**2")
	}

	return nodes, edges, combos, nil
}

func CustomTopoListService(query *response.PaginationQ) ([]*mysqlmanager.Topo_configuration, int, error) {
	tcs := make([]*mysqlmanager.Topo_configuration, 0)

	if mysqlmanager.Global_Mysql == nil {
		err := errors.New("global_mysql is nil **errstackfatal**1")
		return nil, 0, err
	}

	tcdbs, total, err := mysqlmanager.Global_Mysql.QueryTopoConfigurationList(query)
	if err != nil {
		return nil, 0, errors.Wrap(err, "**2")
	}

	for _, tcdb := range tcdbs {
		tc, err := mysqlmanager.Global_Mysql.DBToTopoConfiguration(tcdb)
		if err != nil {
			return nil, 0, errors.Wrap(err, "**2")
		}

		tcs = append(tcs, tc)
	}

	return tcs, total, nil
}

func CreateCustomTopoService(topoconfig *mysqlmanager.Topo_configuration) (int, error) {
	if mysqlmanager.Global_Mysql == nil {
		return -1, errors.New("global_mysql is nil **errstackfatal**1")
	}

	tcdb, err := mysqlmanager.Global_Mysql.TopoConfigurationToDB(topoconfig)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	tcdb.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	tcdb.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	tcdb_id, err := mysqlmanager.Global_Mysql.AddTopoConfiguration(tcdb)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	return tcdb_id, nil
}

func UpdateCustomTopoService(tc *mysqlmanager.Topo_configuration, tcdb_id_old uint) (int, error) {
	if mysqlmanager.Global_Mysql == nil {
		return -1, errors.New("global_mysql is nil **errstackfatal**1")
	}

	tcdb, err := mysqlmanager.Global_Mysql.TopoConfigurationToDB(tc)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	tcdb_old, err := mysqlmanager.Global_Mysql.QuerySingleTopoConfiguration(tcdb_id_old)
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

	tcdb_old_id, err := mysqlmanager.Global_Mysql.AddTopoConfiguration(tcdb_old)
	if err != nil {
		return -1, errors.Wrap(err, "**2")
	}

	return tcdb_old_id, nil
}

func DeleteCustomTopoService(ids []uint) error {
	if mysqlmanager.Global_Mysql == nil {
		return errors.New("global_mysql is nil **errstackfatal**1")
	}
	
	for _, tcid := range ids {
		if err := mysqlmanager.Global_Mysql.DeleteTopoConfiguration(tcid); err != nil {
			return errors.Wrap(err, "**errstack**2")
		}
	}

	return nil
}
