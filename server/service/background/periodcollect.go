package service

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/pkg/errors"
)

func PeriodCollectWorking(batch []string, noderules [][]meta.Filter_rule) {
	graphperiod := conf.Global_config.Topo.Period

	agentmanager.Topo.UpdateMachineList()

	go func(_interval int64, _gdb dao.GraphdbIface, _noderules [][]meta.Filter_rule) {
		for {
			running_agent_num := dao.Global_redis.UpdateTopoRunningAgentList(batch)
			unixtime_now := time.Now().Unix()
			DataProcessWorking(unixtime_now, running_agent_num, _gdb, nil, _noderules)
			time.Sleep(time.Duration(_interval) * time.Second)
		}
	}(graphperiod, dao.Global_GraphDB, noderules)
}

func DataProcessWorking(unixtime int64, agentnum int, graphdb dao.GraphdbIface, tagrules []meta.Tag_rule, noderules [][]meta.Filter_rule) ([]*meta.Node, []*meta.Edge, []map[string]string, error) {
	start := time.Now()

	var nodeTypeWg sync.WaitGroup
	var nodeUuidWg sync.WaitGroup
	var edgeBreakWg sync.WaitGroup
	_unixtime := strconv.Itoa(int(unixtime))

	dataprocesser := processor.CreateDataProcesser()
	nodes, edges, collect_errlist, process_errlist := dataprocesser.ProcessData(agentnum, tagrules, noderules)
	if len(collect_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, "**warn**3") // err top
			agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, collect_errlist[i], agentmanager.Topo.ErrCh, false)
		}
		collect_errlist_string := []string{}
		for _, e := range collect_errlist {
			collect_errlist_string = append(collect_errlist_string, e.Error())
		}
		return nil, nil, nil, errors.Errorf("collect data failed: %+v **10", strings.Join(collect_errlist_string, "/e/"))
	}
	if len(process_errlist) != 0 {
		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, "**warn**14") // err top
			agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, process_errlist[i], agentmanager.Topo.ErrCh, false)
		}
		process_errlist_string := []string{}
		for _, e := range process_errlist {
			process_errlist_string = append(process_errlist_string, e.Error())
		}
		return nil, nil, nil, errors.Errorf("process data failed: %+v **21", strings.Join(process_errlist_string, "/e/"))
	}

	if len(noderules) != 0 {
		combos := make([]map[string]string, 0)

		for _, node := range nodes.Nodes {
			if node.Type == "host" {
				combos = append(combos, map[string]string{
					"id":    node.UUID,
					"label": node.UUID,
				})
			}
		}

		return nodes.Nodes, edges.Edges, combos, nil
	}

	for _, nodesByUUID := range nodes.LookupByUUID {
		nodesbyuuid := nodesByUUID

		nodeUuidWg.Add(1)
		go func(_nodesbyuuid []*meta.Node) {
			defer nodeUuidWg.Done()

			// TODO: 根据插件运行状态agent的数目拆分nodes
			splitnodes := utils.SplitNodesByBreakpoint(_nodesbyuuid, agentnum)
			if splitnodes != nil {
				for _, _nodes := range splitnodes {
					__nodes := _nodes
					nodeTypeWg.Add(1)
					go func(_nodesbytype []*meta.Node) {
						defer nodeTypeWg.Done()

						var cqlIN string

						for _, node := range _nodesbytype {
							_node := node
							if len(_node.Metrics) == 0 {
								cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:'%s', nid:'%s', name:'%s'} set node:'%s')",
									_node.Type, _unixtime, _node.ID, _node.Name, _node.UUID)
							} else {
								cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:'%s', nid:'%s', name:'%s'}) set node:`%s`, node += $metrics",
									_node.Type, _unixtime, _node.ID, _node.Name, _node.UUID)
							}

							err := graphdb.Node_create(_unixtime, _node)
							if err != nil {
								err = errors.Wrapf(err, "create neo4j node failed; %s **warn**2", cqlIN) // err top
								agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
							}
						}
					}(__nodes)
				}
				nodeTypeWg.Wait()
			}
		}(nodesbyuuid)
	}
	nodeUuidWg.Wait()

	splitedges := utils.SplitEdgesByBreakpoint(edges.Edges, agentnum)
	if splitedges != nil {
		for _, _edges := range splitedges {
			__edges := _edges
			edgeBreakWg.Add(1)
			go func(___edges []*meta.Edge) {
				defer edgeBreakWg.Done()

				for _, _edge := range ___edges {
					err := graphdb.Edge_create(_unixtime, _edge)
					if err != nil {
						err = errors.Wrapf(err, "create neo4j edge failed **warn**2") // err top
						agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
					}
				}
			}(__edges)
		}
		edgeBreakWg.Wait()
	}

	elapse := time.Since(start)
	// fmt.Fprintf(agentmanager.Topo.Out, "\033[32mtopo server 数据库写入时间\033[0m: %v\n\n", elapse)
	logger.Info("\033[32mtopo server 数据库写入时间\033[0m: %v\n\n", elapse)

	return nil, nil, nil, nil
}
