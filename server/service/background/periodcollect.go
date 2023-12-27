package service

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"github.com/pkg/errors"
)

func PeriodCollectWorking() {
	graphperiod := conf.Global_config.Topo.Period

	agentmanager.Topo.UpdateMachineList()

	go func(interval int64, gdb dao.GraphdbIface) {
		for {
			runningAgentNum, err := dao.Global_redis.UpdateTopoRunningAgentList()
			if err != nil {
				err = errors.Wrapf(err, "**warn**2") // err top
				agentmanager.Topo.ErrCh <- err
				time.Sleep(5 * time.Second)
				continue
			}

			unixtime_now := time.Now().Unix()
			PeriodProcessWorking(unixtime_now, runningAgentNum, gdb)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}(graphperiod, dao.Global_GraphDB)
}

func PeriodProcessWorking(unixtime int64, agentnum int, graphdb dao.GraphdbIface) {
	start := time.Now()

	var nodeTypeWg sync.WaitGroup
	var nodeUuidWg sync.WaitGroup
	var edgeBreakWg sync.WaitGroup
	_unixtime := strconv.Itoa(int(unixtime))

	dataprocesser := processor.CreateDataProcesser()
	nodes, edges, collect_errlist, process_errlist := dataprocesser.Process_data(agentnum)
	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, "**warn**3") // err top
			agentmanager.Topo.ErrCh <- collect_errlist[i]
		}

		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, "**warn**8") // err top
			agentmanager.Topo.ErrCh <- process_errlist[i]
		}
	}

	for _, nodesByUUID := range nodes.LookupByUUID {
		nodesbyuuid := nodesByUUID

		nodeUuidWg.Add(1)
		go func(_nodesbyuuid []*meta.Node) {
			defer nodeUuidWg.Done()

			// TODO: 根据默认断点数拆分nodes
			for _, _nodes := range utils.SplitNodesByBreakpoint(_nodesbyuuid, agentnum) {
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
							agentmanager.Topo.ErrCh <- err
						}
					}
				}(__nodes)
			}
			nodeTypeWg.Wait()

		}(nodesbyuuid)
	}
	nodeUuidWg.Wait()

	for _, _edges := range utils.SplitEdgesByBreakpoint(edges.Edges, agentnum) {
		__edges := _edges
		edgeBreakWg.Add(1)
		go func(___edges []*meta.Edge) {
			defer edgeBreakWg.Done()

			for _, _edge := range ___edges {
				err := graphdb.Edge_create(_unixtime, _edge)
				if err != nil {
					err = errors.Wrapf(err, "create neo4j edge failed **warn**2") // err top
					agentmanager.Topo.ErrCh <- err
				}
			}
		}(__edges)
	}

	edgeBreakWg.Wait()

	elapse := time.Since(start)
	fmt.Fprintf(agentmanager.Topo.Out, "\033[32mtopo server 数据库写入时间\033[0m: %v\n\n", elapse)
}
