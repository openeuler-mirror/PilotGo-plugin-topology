package custom

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/redismanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/generator"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/graph"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"

	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/pkg/errors"
)

func RunCustomTopoService(tcid uint) (*graph.Nodes, *graph.Edges, []map[string]string, error) {
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

	topogenerator := generator.CreateTopoGenerator(tc.TagRules, tc.NodeRules)
	nodes, edges, collect_errlist, process_errlist := topogenerator.ProcessingData(running_agent_num)
	if len(collect_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, "**errstack**3") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, collect_errlist[i], false)
		}
		collect_errlist_string := []string{}
		for _, e := range collect_errlist {
			collect_errlist_string = append(collect_errlist_string, e.Error())
		}
		return nil, nil, nil, errors.Errorf("collect data failed: %+v **errstack**10", strings.Join(collect_errlist_string, "/e/"))
	}
	if len(process_errlist) != 0 {
		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, "**errstack**14") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, process_errlist[i], false)
		}
		process_errlist_string := []string{}
		for _, e := range process_errlist {
			process_errlist_string = append(process_errlist_string, e.Error())
		}
		return nil, nil, nil, errors.Errorf("process data failed: %+v **errstack**21", strings.Join(process_errlist_string, "/e/"))
	}
	if nodes == nil || edges == nil {
		err := errors.New("nodes or edges is nil **errstack**24") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		return nil, nil, nil, err
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
