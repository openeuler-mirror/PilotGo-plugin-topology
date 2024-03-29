package dao

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
)

type Neo4jClient struct {
	addr     string
	username string
	password string
	DB       string
	Driver   neo4j.Driver
}

var Global_Neo4j *Neo4jClient

func Neo4jInit(url, user, pass, db string) *Neo4jClient {
	n := &Neo4jClient{
		addr:     url,
		username: user,
		password: pass,
		DB:       db,
	}

	driver, err := neo4j.NewDriver(n.addr, neo4j.BasicAuth(n.username, n.password, ""), func(config *neo4j.Config) {
		config.MaxTransactionRetryTime = 30 * time.Second
		config.MaxConnectionPoolSize = 50
		config.MaxConnectionLifetime = 1 * time.Hour
	})
	if err != nil {
		err := errors.Errorf("create neo4j driver failed: %s **errstackfatal**2", err.Error()) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, true)
	}

	n.Driver = driver
	return n
}

func (n *Neo4jClient) Node_create(unixtime string, node *meta.Node) error {
	var cqlIN string

	if Global_Neo4j == nil {
		return errors.New("neo4j client not init **errstack**1")
	}

	if len(node.Metrics) == 0 {
		cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:'%s', nid:'%s', name:'%s', layoutattr:'%s', comboid:'%s'} set node:`%s`)",
			node.Type, unixtime, node.ID, node.Name, node.LayoutAttr, node.ComboId, node.UUID)
	} else {
		cqlIN = fmt.Sprintf("create (node:`%s` {unixtime:'%s', nid:'%s', name:'%s', layoutattr:'%s', comboid:'%s'}) set node:`%s`, node += $metrics",
			node.Type, unixtime, node.ID, node.Name, node.LayoutAttr, node.ComboId, node.UUID)
	}

	params := map[string]interface{}{
		"metrics": node.Metrics,
	}

	session := n.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: n.DB})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cqlIN, params)
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})
	if err != nil {
		err = errors.Errorf("neo4j writetransaction failed: %s, %s **errstack**9", err.Error(), cqlIN)
		return err
	}

	return nil
}

func (n *Neo4jClient) Edge_create(unixtime string, edge *meta.Edge) error {
	var cqlIN string

	if Global_Neo4j == nil {
		return errors.New("neo4j client not init **errstack**1")
	}

	if len(edge.Metrics) == 0 {
		cqlIN = fmt.Sprintf("match (src {unixtime:'%s', nid:'%s'}), (dst {unixtime:'%s', nid:'%s'}) create (src)-[r:`%s` {unixtime:'%s', rid:'%s', dir:'%s', src:'%s', dst:'%s'}]->(dst)",
			unixtime, edge.Src, unixtime, edge.Dst, edge.Type, unixtime, edge.ID, edge.Dir, edge.Src, edge.Dst)
	} else {
		cqlIN = fmt.Sprintf("match (src {unixtime:'%s', nid:'%s'}), (dst {unixtime:'%s', nid:'%s'}) create (src)-[r:`%s` {unixtime:'%s', rid:'%s', dir:'%s', src:'%s', dst:'%s'}]->(dst) set r += $metrics",
			unixtime, edge.Src, unixtime, edge.Dst, edge.Type, unixtime, edge.ID, edge.Dir, edge.Src, edge.Dst)
	}

	params := map[string]interface{}{
		"metrics": edge.Metrics,
	}

	session := n.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: n.DB})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cqlIN, params)
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})
	if err != nil {
		err = errors.Errorf("neo4j writetransaction failed: %s, %s **errstack**9", err.Error(), cqlIN)
		return err
	}

	return nil
}

func (n *Neo4jClient) Timestamps_query() ([]string, error) {
	var cqlOUT string
	var varia string
	cqlOUT = "match (n:host) return collect(distinct n.unixtime) as times"
	varia = "times"
	var list []string

	if Global_Neo4j == nil {
		return nil, errors.New("neo4j client not init **errstack**1")
	}

	session := n.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: n.DB})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cqlOUT, nil)
		if err != nil {
			err = errors.Errorf("neo4j query failed: %s, %s **errstack**2", err.Error(), cqlOUT)
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
			err = errors.Errorf("iterate result failed: %s, %s **errstack**1", err.Error(), cqlOUT)
			return nil, err
		}

		return list, result.Err()
	})

	if err != nil {
		err = errors.Errorf("query Readtransaction error: %s, %s **errstack**26", err.Error(), cqlOUT)
		return nil, err
	}

	sort.Strings(list)

	return list, nil
}

func (n *Neo4jClient) SingleHost_node_query(uuid string, unixtime string) ([]*meta.Node, error) {
	var cqlOUT string
	var varia string
	cqlOUT = fmt.Sprintf("match (nodes:`%s`) where nodes.unixtime='%s' return nodes", uuid, unixtime)
	varia = "nodes"

	if Global_Neo4j == nil {
		return nil, errors.New("neo4j client not init **errstack**1")
	}

	return n.Node_query(cqlOUT, varia)
}

func (n *Neo4jClient) MultiHost_node_query(unixtime string) ([]*meta.Node, error) {
	var cqlOUT string
	var varia string
	cqlOUT = fmt.Sprintf("match (nodes) where nodes.unixtime='%s' return nodes", unixtime)
	varia = "nodes"

	if Global_Neo4j == nil {
		return nil, errors.New("neo4j client not init **errstack**1")
	}

	return n.Node_query(cqlOUT, varia)
}

func (n *Neo4jClient) Node_query(cypher string, varia string) ([]*meta.Node, error) {
	var list []*meta.Node
	session := n.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: n.DB})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			err = errors.Errorf("neo4j query failed: %s, %s **errstack**2", err.Error(), cypher)
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
			err = errors.Errorf("iterate node result failed: %s, %s **errstack**1", err.Error(), cypher)
			return nil, err
		}

		return list, result.Err()
	})

	if err != nil {
		err = errors.Errorf("node Readtransaction error: %s, %s **errstack**24", err.Error(), cypher)
	}
	return list, err
}

func (n *Neo4jClient) MultiHost_relation_query(unixtime string) ([]*meta.Edge, error) {
	var cqlOUT string
	var varia string
	cqlOUT = fmt.Sprintf("match ()-[relas]->() where relas.unixtime='%s' return relas", unixtime)
	varia = "relas"

	if Global_Neo4j == nil {
		return nil, errors.New("neo4j client not init **errstack**1")
	}

	return n.Relation_query(cqlOUT, varia)
}

func (n *Neo4jClient) Relation_query(cypher string, varia string) ([]*meta.Edge, error) {
	var list []*meta.Edge
	session := n.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: n.DB})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			err = errors.Errorf("RelationshipQuery failed: %s, %s **errstack**2", err.Error(), cypher)
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
			err = errors.Errorf("iterate relation result failed: %s, %s **errstack**1", err.Error(), cypher)
			return nil, err
		}
		return list, result.Err()
	})

	if err != nil {
		err = errors.Errorf("relation Readtransaction error: %s, %s **errstack**22", err.Error(), cypher)
	}
	return list, err
}

func (n *Neo4jClient) ClearExpiredData(retention int64) {
	if Global_Neo4j == nil {
		err := errors.New("neo4j client not init **errstack**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		return
	}

	current := time.Now()
	timepoint := current.Add(-time.Duration(retention) * time.Hour).Unix()
	cqlIN := `match (n) where n.unixtime < $timepoint detach delete n`
	params := map[string]interface{}{
		"timepoint": strconv.Itoa(int(timepoint)),
	}

	session := n.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: n.DB})
	defer session.Close()

	result, err := session.Run(cqlIN, params)
	if err != nil {
		err = errors.Errorf("ClearExpiredData failed: %s, %s **warn**1", err.Error(), cqlIN) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		return
	}

	summary, err := result.Consume()
	if err != nil {
		err = errors.Errorf("failed to consume ClearExpiredData result: %s, %s **warn**2", err.Error(), cqlIN) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		return
	}

	err = errors.Errorf("delete %d nodes **debug**0", summary.Counters().NodesDeleted()) // err top
	agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
}
