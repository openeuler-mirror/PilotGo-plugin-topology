package dao

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
)

type Neo4jclient struct {
	addr     string
	username string
	password string
	DB       string

	driver neo4j.Driver
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

func (n *Neo4jclient) Close_driver() error {
	return n.driver.Close()
}

func (n *Neo4jclient) Entity_create(cypher string, params map[string]interface{}) error {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: n.DB})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, params)
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})

	if err != nil {
		err = errors.Errorf("neo4j writetransaction failed: %s **9", err.Error())
		return err
	}

	return nil
}

func (n *Neo4jclient) Node_query(cypher string) ([]neo4j.Node, error) {
	var list []neo4j.Node
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: n.DB})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			fmt.Println("NodeQuery failed: ", err)
			return nil, err
		}

		for result.Next() {
			record := result.Record()
			if value, ok := record.Get("n"); ok {
				node := value.(neo4j.Node)
				list = append(list, node)
			}
		}
		if err = result.Err(); err != nil {
			fmt.Println("iterate node result failed: ", err)
			return nil, err
		}

		return list, result.Err()
	})

	if err != nil {
		fmt.Println("node Readtransaction error:", err)
	}
	return list, err
}

func (n *Neo4jclient) Relation_query(cypher string) ([]neo4j.Relationship, error) {
	var list []neo4j.Relationship
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: n.DB})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			fmt.Println("RelationshipQuery failed: ", err)
			return nil, err
		}
		for result.Next() {
			record := result.Record()
			if value, ok := record.Get("r"); ok {
				relationship := value.(neo4j.Relationship)
				list = append(list, relationship)
			}
		}
		if err = result.Err(); err != nil {
			fmt.Println("iterate relation result failed: ", err)
			return nil, err
		}
		return list, result.Err()
	})

	if err != nil {
		fmt.Println("relation Readtransaction error:", err)
	}
	return list, err
}

func PeriodProcessNeo4j(unixtime int64, agentnum int) {
	start := time.Now()

	var nodeTypeWg sync.WaitGroup
	var nodeUuidWg sync.WaitGroup
	var edgeBreakWg sync.WaitGroup
	_unixtime := unixtime

	_neo4j := CreateNeo4j(conf.Global_config.Neo4j.Addr, conf.Global_config.Neo4j.Username, conf.Global_config.Neo4j.Password, conf.Global_config.Neo4j.DB)

	d, err := _neo4j.Create_driver()
	if err != nil {
		err := errors.Errorf("create neo4j driver failed: %s **fatal**2", err.Error()) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.Errmu.Lock()
		agentmanager.Topo.ErrCond.Wait()
		agentmanager.Topo.Errmu.Unlock()
		close(agentmanager.Topo.ErrCh)
		os.Exit(1)
	}
	_neo4j.driver = d
	defer _neo4j.Close_driver()

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
							cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:%d, nid:'%s', name:'%s'} set node:'%s')",
								_node.Type, _unixtime, _node.ID, _node.Name, _node.UUID)
						} else {
							cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:%d, nid:'%s', name:'%s'}) set node:`%s`, node += $metrics",
								_node.Type, _unixtime, _node.ID, _node.Name, _node.UUID)
						}

						params := map[string]interface{}{
							"metrics": _node.Metrics,
						}

						err := _neo4j.Entity_create(cqlIN, params)
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
					cqlIN = fmt.Sprintf("match (src {unixtime:%d, nid:'%s'}), (dst {unixtime:%d, nid:'%s'}) create (src)-[r:`%s` {unixtime:%d, rid:'%s', dir:%t}]->(dst)",
						_unixtime, _edge.Src, _unixtime, _edge.Dst, _edge.Type, _unixtime, _edge.ID, _edge.Dir)
				} else {
					cqlIN = fmt.Sprintf("match (src {unixtime:%d, nid:'%s'}), (dst {unixtime:%d, nid:'%s'}) create (src)-[r:`%s` {unixtime:%d, rid:'%s', dir:%t}]->(dst) set r += $metrics",
						_unixtime, _edge.Src, _unixtime, _edge.Dst, _edge.Type, _unixtime, _edge.ID, _edge.Dir)
				}
				params := map[string]interface{}{
					"metrics": _edge.Metrics,
				}

				err := _neo4j.Entity_create(cqlIN, params)
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
