package service

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"github.com/pkg/errors"
)

func PeriodProcessNeo4j(unixtime int64, agentnum int) {
	start := time.Now()

	var nodeTypeWg sync.WaitGroup
	var nodeUuidWg sync.WaitGroup
	var edgeBreakWg sync.WaitGroup
	_unixtime := strconv.Itoa(int(unixtime))

	dataprocesser := processor.CreateDataProcesser()
	nodes, edges, collect_errlist, process_errlist := dataprocesser.Process_data()
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

	// TODO: 临时获取运行状态agent的数目
	_agentnum := agentmanager.Topo.GetRunningAgentNumber()
	if _agentnum <= 0 {
		err := errors.New("no running agent **warn**2") // err top
		agentmanager.Topo.ErrCh <- err
		return
	}

	for _, nodesByUUID := range nodes.LookupByUUID {
		nodesbyuuid := nodesByUUID

		nodeUuidWg.Add(1)
		go func(_nodesbyuuid []*meta.Node) {
			defer nodeUuidWg.Done()

			// TODO: 根据默认断点数拆分nodes
			for _, _nodes := range utils.SplitNodesByBreakpoint(_nodesbyuuid, 10) {
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

						params := map[string]interface{}{
							"metrics": _node.Metrics,
						}

						err := dao.Neo4j.Entity_create(cqlIN, params)
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

	for _, _edges := range utils.SplitEdgesByBreakpoint(edges.Edges, int(_agentnum)) {
		__edges := _edges
		edgeBreakWg.Add(1)
		go func(___edges []*meta.Edge) {
			defer edgeBreakWg.Done()

			var cqlIN string

			for _, _edge := range ___edges {
				if len(_edge.Metrics) == 0 {
					cqlIN = fmt.Sprintf("match (src {unixtime:'%s', nid:'%s'}), (dst {unixtime:'%s', nid:'%s'}) create (src)-[r:`%s` {unixtime:'%s', rid:'%s', dir:'%s', src:'%s', dst:'%s'}]->(dst)",
						_unixtime, _edge.Src, _unixtime, _edge.Dst, _edge.Type, _unixtime, _edge.ID, _edge.Dir, _edge.Src, _edge.Dst)
				} else {
					cqlIN = fmt.Sprintf("match (src {unixtime:'%s', nid:'%s'}), (dst {unixtime:'%s', nid:'%s'}) create (src)-[r:`%s` {unixtime:'%s', rid:'%s', dir:'%s', src:'%s', dst:'%s'}]->(dst) set r += $metrics",
						_unixtime, _edge.Src, _unixtime, _edge.Dst, _edge.Type, _unixtime, _edge.ID, _edge.Dir, _edge.Src, _edge.Dst)
				}
				params := map[string]interface{}{
					"metrics": _edge.Metrics,
				}

				err := dao.Neo4j.Entity_create(cqlIN, params)
				if err != nil {
					err = errors.Wrapf(err, "create neo4j edge failed **warn**2") // err top
					agentmanager.Topo.ErrCh <- err
				}
			}
		}(__edges)
	}

	edgeBreakWg.Wait()

	elapse := time.Since(start)
	fmt.Fprintf(agentmanager.Topo.Out, "\033[32mtopo server 数据库写入时间\033[0m: %v\n", elapse)
}
