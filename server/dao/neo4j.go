package dao

import (
	"fmt"
	"os"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
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
		err = errors.Errorf("neo4j writetransaction failed: %s", err.Error())
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

func PeriodProcessNeo4j(unixtime int64) {
	var cqlIN string

	_neo4j := CreateNeo4j(conf.Global_config.Neo4j.Addr, conf.Global_config.Neo4j.Username, conf.Global_config.Neo4j.Password, conf.Global_config.Neo4j.DB)

	d, err := _neo4j.Create_driver()
	if err != nil {
		err := errors.Errorf("create neo4j driver failed: %s **fatal**2", err.Error()) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.ErrGroup.Add(1)
		agentmanager.Topo.ErrGroup.Wait()
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

	for _, node := range nodes.Nodes {
		if len(node.Metrics) == 0 {
			cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:%d, nid:'%s', name:'%s'} set node:'%s')",
				node.Type, unixtime, node.ID, node.Name, node.UUID)
		} else {
			cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:%d, nid:'%s', name:'%s'}) set node:`%s`, node += $metrics",
				node.Type, unixtime, node.ID, node.Name, node.UUID)
		}

		params := map[string]interface{}{
			"metrics": node.Metrics,
		}

		err := _neo4j.Entity_create(cqlIN, params)
		if err != nil {
			err = errors.Wrapf(err, "create neo4j node failed **warn**2") // err top
			agentmanager.Topo.ErrCh <- err
		}
	}

	for _, edge := range edges.Edges {
		if len(edge.Metrics) == 0 {
			cqlIN = fmt.Sprintf("match (src {unixtime:%d, nid:'%s'}), (dst {unixtime:%d, nid:'%s'}) create (src)-[r:`%s` {unixtime:%d, rid:'%s', dir:%t}]->(dst)",
				unixtime, edge.Src, unixtime, edge.Dst, edge.Type, unixtime, edge.ID, edge.Dir)
		} else {
			cqlIN = fmt.Sprintf("match (src {unixtime:%d, nid:'%s'}), (dst {unixtime:%d, nid:'%s'}) create (src)-[r:`%s` {unixtime:%d, rid:'%s', dir:%t}]->(dst) set r += $metrics",
				unixtime, edge.Src, unixtime, edge.Dst, edge.Type, unixtime, edge.ID, edge.Dir)
		}
		params := map[string]interface{}{
			"metrics": edge.Metrics,
		}

		err := _neo4j.Entity_create(cqlIN, params)
		if err != nil {
			err = errors.Wrapf(err, "create neo4j edge failed **warn**2") // err top
			agentmanager.Topo.ErrCh <- err
		}
	}
}
