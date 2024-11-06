package custom

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/db/redismanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/generator"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/graph"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/pluginclient"

	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/pkg/errors"
)

func RunCustomTopoService(tcid uint) (*graph.Nodes, *graph.Edges, []map[string]string, error, bool) {
	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil")
		return nil, nil, nil, err, true
	}
	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil")
		return nil, nil, nil, err, true
	}
	if redismanager.Global_Redis == nil {
		err := errors.New("global_redis is nil")
		return nil, nil, nil, err, true
	}
	if mysqlmanager.Global_Mysql == nil {
		err := errors.New("global_mysql is nil")
		return nil, nil, nil, err, true
	}

	tcdb, err := mysqlmanager.Global_Mysql.QuerySingleTopoConfiguration(tcid)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, " "), false
	}

	tc, err := mysqlmanager.Global_Mysql.DBToTopoConfiguration(tcdb)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, " "), false
	}

	machine_uuids, err := pluginclient.Global_Client.BatchUUIDList(strconv.Itoa(int(tc.BatchId)))
	if err != nil {
		return nil, nil, nil, errors.New(err.Error()), false
	}

	// ctxv := context.WithValue(agentmanager.Topo.Tctx, "custom_name", "pilotgo-topo")

	agentmanager.Global_AgentManager.UpdateMachineList()

	redismanager.Global_Redis.ActiveHeartbeatDetection(machine_uuids)
	running_agent_num := redismanager.Global_Redis.UpdateTopoRunningAgentList(machine_uuids, true)
	if running_agent_num == 0 {
		return nil, nil, nil, errors.Errorf("no running agent for custom id %d", tc.ID), false
	} else if running_agent_num == -1 {
		return nil, nil, nil, errors.New("redis client not init"), false
	}

	topogenerator := generator.CreateTopoGenerator(tc.TagRules, tc.NodeRules)
	nodes, edges, collect_errlist, process_errlist := topogenerator.ProcessingData(running_agent_num)
	if len(collect_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, " ")
			global.ERManager.ErrorTransmit("error", collect_errlist[i], false, true)
		}
		collect_errlist_string := []string{}
		for _, e := range collect_errlist {
			collect_errlist_string = append(collect_errlist_string, e.Error())
		}
		return nil, nil, nil, errors.Errorf("collect data failed: %+v", strings.Join(collect_errlist_string, "/e/")), false
	}
	if len(process_errlist) != 0 {
		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, " ")
			global.ERManager.ErrorTransmit("error", process_errlist[i], false, true)
		}
		process_errlist_string := []string{}
		for _, e := range process_errlist {
			process_errlist_string = append(process_errlist_string, e.Error())
		}
		return nil, nil, nil, errors.Errorf("process data failed: %+v", strings.Join(process_errlist_string, "/e/")), false
	}
	if nodes == nil || edges == nil {
		err := errors.New("nodes or edges is nil")
		global.ERManager.ErrorTransmit("error", err, false, true)
		return nil, nil, nil, err, false
	}

	combos := make([]map[string]string, 0)
	for _, node := range nodes.Nodes {
		if node.Type == "host" {
			combos = append(combos, map[string]string{
				"id":    node.UUID,
				"label": fmt.Sprintf("%s/%s", node.Metrics["Hostname"], strings.Split(node.ID, global.NODE_CONNECTOR)[len(strings.Split(node.ID, global.NODE_CONNECTOR))-1]),
			})
		}
	}

	return nodes, edges, combos, nil, false
}

func CustomTopoListService(query *response.PaginationQ) ([]*mysqlmanager.Topo_configuration, int, error, bool) {
	tcs := make([]*mysqlmanager.Topo_configuration, 0)

	if mysqlmanager.Global_Mysql == nil {
		err := errors.New("global_mysql is nil")
		return nil, 0, err, true
	}

	tcdbs, total, err := mysqlmanager.Global_Mysql.QueryTopoConfigurationList(query)
	if err != nil {
		return nil, 0, errors.Wrap(err, " "), false
	}

	for _, tcdb := range tcdbs {
		tc, err := mysqlmanager.Global_Mysql.DBToTopoConfiguration(tcdb)
		if err != nil {
			return nil, 0, errors.Wrap(err, " "), false
		}

		tcs = append(tcs, tc)
	}

	return tcs, total, nil, false
}

func CreateCustomTopoService(topoconfig *mysqlmanager.Topo_configuration) (int, error, bool) {
	if mysqlmanager.Global_Mysql == nil {
		return -1, errors.New("global_mysql is nil"), true
	}

	tcdb, err := mysqlmanager.Global_Mysql.TopoConfigurationToDB(topoconfig)
	if err != nil {
		return -1, errors.Wrap(err, " "), false
	}

	tcdb.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	tcdb.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	tcdb_id, err := mysqlmanager.Global_Mysql.AddTopoConfiguration(tcdb)
	if err != nil {
		return -1, errors.Wrap(err, " "), false
	}

	return tcdb_id, nil, false
}

func UpdateCustomTopoService(tc *mysqlmanager.Topo_configuration, tcdb_id_old uint) (int, error, bool) {
	if mysqlmanager.Global_Mysql == nil {
		return -1, errors.New("global_mysql is nil"), true
	}

	tcdb, err := mysqlmanager.Global_Mysql.TopoConfigurationToDB(tc)
	if err != nil {
		return -1, errors.Wrap(err, " "), false
	}

	tcdb_old, err := mysqlmanager.Global_Mysql.QuerySingleTopoConfiguration(tcdb_id_old)
	if err != nil {
		return -1, errors.Wrap(err, " "), false
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
		return -1, errors.Wrap(err, " "), false
	}

	return tcdb_old_id, nil, false
}

func DeleteCustomTopoService(ids []uint) (error, bool) {
	if mysqlmanager.Global_Mysql == nil {
		return errors.New("global_mysql is nil"), true
	}

	for _, tcid := range ids {
		if err := mysqlmanager.Global_Mysql.DeleteTopoConfiguration(tcid); err != nil {
			return errors.Wrap(err, " "), false
		}
	}

	return nil, false
}
