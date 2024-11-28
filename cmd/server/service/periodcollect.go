/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package service

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/db/graphmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/db/redismanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/generator"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/graph"
	"github.com/pkg/errors"
)

func InitPeriodCollectWorking(batch []string, noderules [][]mysqlmanager.Filter_rule) {
	if graphmanager.Global_GraphDB == nil {
		err := errors.New("global_graphdb is nil")
		global.ERManager.ErrorTransmit("service", "info", err, false, false)
		return
	}

	graphperiod := conf.Global_Config.Neo4j.Period

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil")
		global.ERManager.ErrorTransmit("service", "error", err, true, true)
		return
	}

	if redismanager.Global_Redis == nil {
		err := errors.New("Global_Redis is nil")
		global.ERManager.ErrorTransmit("service", "error", err, true, true)
		return
	}

	agentmanager.Global_AgentManager.UpdateMachineList()

	global.ERManager.Wg.Add(1)
	go func(_interval int64, _gdb graphmanager.GraphdbIface, _noderules [][]mysqlmanager.Filter_rule) {
		defer global.ERManager.Wg.Done()
		for {
			select {
			case <-global.ERManager.GoCancelCtx.Done():
				global.ERManager.ErrorTransmit("service", "info", errors.New("cancelCtx is done, exit period collect goroutine"), false, false)
				return
			default:
				redismanager.Global_Redis.ActiveHeartbeatDetection(batch)
				running_agent_num := redismanager.Global_Redis.UpdateTopoRunningAgentList(batch, false)
				unixtime_now := time.Now().Unix()
				DataProcessWorking(unixtime_now, running_agent_num, _gdb, nil, _noderules)
				time.Sleep(time.Duration(_interval) * time.Second)
			}
		}
	}(graphperiod, graphmanager.Global_GraphDB, noderules)
}

func DataProcessWorking(unixtime int64, agentnum int, graphdb graphmanager.GraphdbIface, tagrules []mysqlmanager.Tag_rule, noderules [][]mysqlmanager.Filter_rule) ([]*graph.Node, []*graph.Edge, []map[string]string, error) {
	var nodeTypeWg sync.WaitGroup
	var nodeUuidWg sync.WaitGroup
	var edgeBreakWg sync.WaitGroup
	_unixtime := strconv.Itoa(int(unixtime))

	topogenerator := generator.CreateTopoGenerator(tagrules, noderules)
	nodes, edges, collect_errlist, process_errlist := topogenerator.ProcessingData(agentnum)
	if len(collect_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, " ")
			global.ERManager.ErrorTransmit("service", "error", collect_errlist[i], false, true)
		}
		collect_errlist_string := []string{}
		for _, e := range collect_errlist {
			collect_errlist_string = append(collect_errlist_string, e.Error())
		}
		return nil, nil, nil, errors.Errorf("collect data failed: %+v", strings.Join(collect_errlist_string, "/e/"))
	}
	if len(process_errlist) != 0 {
		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, " ")
			global.ERManager.ErrorTransmit("service", "error", process_errlist[i], false, true)
		}
		process_errlist_string := []string{}
		for _, e := range process_errlist {
			process_errlist_string = append(process_errlist_string, e.Error())
		}
		return nil, nil, nil, errors.Errorf("process data failed: %+v", strings.Join(process_errlist_string, "/e/"))
	}
	if nodes == nil || edges == nil {
		err := errors.New("nodes or edges is nil")
		global.ERManager.ErrorTransmit("service", "error", err, false, true)
		return nil, nil, nil, err
	}

	if graphmanager.Global_GraphDB == nil {
		err := errors.New("Global_GraphDB is nil")
		global.ERManager.ErrorTransmit("service", "error", err, true, true)
		return nil, nil, nil, err
	}

	start := time.Now()

	for _, nodesByUUID := range nodes.LookupByUUID {
		nodesbyuuid := nodesByUUID

		nodeUuidWg.Add(1)
		go func(_nodesbyuuid []*graph.Node) {
			defer nodeUuidWg.Done()

			// TODO: 根据插件运行状态agent的数目拆分nodes
			splitnodes := SplitNodesByBreakpoint(_nodesbyuuid, agentnum)
			if splitnodes != nil {
				for _, _nodes := range splitnodes {
					__nodes := _nodes
					nodeTypeWg.Add(1)
					go func(_nodesbytype []*graph.Node) {
						defer nodeTypeWg.Done()

						var cqlIN string

						for _, node := range _nodesbytype {
							_node := node
							if len(_node.Metrics) == 0 {
								cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:'%s', nid:'%s', name:'%s', layoutattr:'%s', comboid:'%s'} set node:`%s`)",
									_node.Type, _unixtime, _node.ID, _node.Name, _node.LayoutAttr, _node.ComboId, _node.UUID)
							} else {
								cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:'%s', nid:'%s', name:'%s', layoutattr:'%s', comboid:'%s'}) set node:`%s`, node += $metrics",
									_node.Type, _unixtime, _node.ID, _node.Name, _node.LayoutAttr, _node.ComboId, _node.UUID)
							}

							err := graphmanager.Global_GraphDB.Node_create(_unixtime, _node)
							if err != nil {
								err = errors.Wrapf(err, "create neo4j node failed; %s", cqlIN)
								global.ERManager.ErrorTransmit("service", "error", err, false, true)
							}
						}
					}(__nodes)
				}
				nodeTypeWg.Wait()
			}
		}(nodesbyuuid)
	}
	nodeUuidWg.Wait()

	splitedges := SplitEdgesByBreakpoint(edges.Edges, agentnum)
	if splitedges != nil {
		for _, _edges := range splitedges {
			__edges := _edges
			edgeBreakWg.Add(1)
			go func(___edges []*graph.Edge) {
				defer edgeBreakWg.Done()

				for _, _edge := range ___edges {
					err := graphmanager.Global_GraphDB.Edge_create(_unixtime, _edge)
					if err != nil {
						err = errors.Wrapf(err, "create neo4j edge failed")
						global.ERManager.ErrorTransmit("service", "error", err, false, true)
					}
				}
			}(__edges)
		}
		edgeBreakWg.Wait()
	}

	elapse := time.Since(start)
	global.ERManager.ErrorTransmit("service", "info", errors.Errorf("图数据库写入时间: %v\n\n", elapse), false, false)
	return nil, nil, nil, nil
}

func SplitEdgesByBreakpoint(arr []*graph.Edge, n int) [][]*graph.Edge {
	length := len(arr)
	if length == 0 {
		return nil
	}

	size := length / n
	result := make([][]*graph.Edge, n)

	for i := 0; i < n; i++ {
		start := i * size
		end := (i + 1) * size

		if end > length {
			end = length
		}

		result = append(result, arr[start:end])
	}

	return result
}

func SplitNodesByBreakpoint(arr []*graph.Node, n int) [][]*graph.Node {
	length := len(arr)
	if length == 0 {
		return nil
	}

	size := length / n
	result := make([][]*graph.Node, n)

	for i := 0; i < n; i++ {
		start := i * size
		end := (i + 1) * size

		if end > length {
			end = length
		}

		result = append(result, arr[start:end])
	}

	return result
}
