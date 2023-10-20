package dao

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
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

func (n *Neo4jclient) Entity_create(cypher string) error {

	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: n.DB})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})

	return err
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
