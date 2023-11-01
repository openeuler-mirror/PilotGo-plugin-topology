package dao

import (
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
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
