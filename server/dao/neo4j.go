package dao

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
)

var Neo4j *Neo4jclient

type Neo4jclient struct {
	addr     string
	username string
	password string
	DB       string
}

func CreateNeo4j(url, user, pass, db string) *Neo4jclient {
	return &Neo4jclient{
		addr:     url,
		username: user,
		password: pass,
		DB:       db,
	}
}

func (n *Neo4jclient) Create_driver() (neo4j.Driver, error) {
	return neo4j.NewDriver(n.addr, neo4j.BasicAuth(n.username, n.password, ""))
}

func (n *Neo4jclient) Close_driver(driver neo4j.Driver) error {
	return driver.Close()
}

func (n *Neo4jclient) Entity_create(cypher string, params map[string]interface{}, driver neo4j.Driver) error {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: n.DB})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, params)
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})

	if err != nil {
		err = errors.Errorf("neo4j writetransaction failed: %s, %s **9", err.Error(), cypher)
		return err
	}

	return nil
}

func (n *Neo4jclient) General_query(cypher string, varia string, driver neo4j.Driver) ([]string, error) {
	var list []string
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: n.DB})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			err = errors.Errorf("neo4j query failed: %s, %s **2", err.Error(), cypher)
			return nil, err
		}

		for result.Next() {
			record := result.Record()
			if value, ok := record.Get(varia); ok {
				_value := value.([]interface{})
				for _, v := range _value {
					list = append(list, v.(string))
				}
			}
		}

		if err := result.Err(); err != nil {
			err = errors.Errorf("iterate result failed: %s, %s **1", err.Error(), cypher)
			return nil, err
		}

		return list, result.Err()
	})

	if err != nil {
		err = errors.Errorf("query Readtransaction error: %s, %s **26", err.Error(), cypher)
	}

	return list, nil
}

func (n *Neo4jclient) Node_query(cypher string, varia string, driver neo4j.Driver) ([]*meta.Node, error) {
	var list []*meta.Node
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: n.DB})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			err = errors.Errorf("neo4j query failed: %s, %s **2", err.Error(), cypher)
			return nil, err
		}

		for result.Next() {
			record := result.Record()
			if value, ok := record.Get(varia); ok {
				neo4jnode := value.(neo4j.Node)
				toponode := utils.Neo4jnodeToToponode(neo4jnode)
				list = append(list, toponode)
			}
		}
		if err := result.Err(); err != nil {
			err = errors.Errorf("iterate node result failed: %s, %s **1", err.Error(), cypher)
			return nil, err
		}

		return list, result.Err()
	})

	if err != nil {
		err = errors.Errorf("node Readtransaction error: %s, %s **24", err.Error(), cypher)
	}
	return list, err
}

func (n *Neo4jclient) Relation_query(cypher string, varia string, driver neo4j.Driver) ([]*meta.Edge, error) {
	var list []*meta.Edge
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: n.DB})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			err = errors.Errorf("RelationshipQuery failed: %s, %s **2", err.Error(), cypher)
			return nil, err
		}
		for result.Next() {
			record := result.Record()
			if value, ok := record.Get(varia); ok {
				relationship := value.(neo4j.Relationship)
				toporelation := utils.Neo4jrelaToToporela(relationship)
				list = append(list, toporelation)
			}
		}
		if err = result.Err(); err != nil {
			err = errors.Errorf("iterate relation result failed: %s, %s **1", err.Error(), cypher)
			return nil, err
		}
		return list, result.Err()
	})

	if err != nil {
		err = errors.Errorf("relation Readtransaction error: %s, %s **22", err.Error(), cypher)
	}
	return list, err
}

func PeriodProcessNeo4j(unixtime int64, agentnum int) {
	start := time.Now()

	var nodeTypeWg sync.WaitGroup
	var nodeUuidWg sync.WaitGroup
	var edgeBreakWg sync.WaitGroup
	_unixtime := strconv.Itoa(int(unixtime))

	dri, err := Neo4j.Create_driver()
	if err != nil {
		err := errors.Errorf("create neo4j driver failed: %s **fatal**2", err.Error()) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.Errmu.Lock()
		agentmanager.Topo.ErrCond.Wait()
		agentmanager.Topo.Errmu.Unlock()
		close(agentmanager.Topo.ErrCh)
		os.Exit(1)
	}
	defer Neo4j.Close_driver(dri)

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

						err := Neo4j.Entity_create(cqlIN, params, dri)
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

				err := Neo4j.Entity_create(cqlIN, params, dri)
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
