package service

import (
	"fmt"
	"os"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"github.com/pkg/errors"
)

func SingleHostTreeService(uuid string) (*TreeTopoNode, error) {
	var latest string
	var cqlOUT string
	var treerootnode *TreeTopoNode
	single_nodes := make([]*meta.Node, 0)
	single_nodes_map := make(map[int64]*meta.Node)
	treenodes_process := make([]*TreeTopoNode, 0)
	treenodes_net := make([]*TreeTopoNode, 0)
	nodes_type_map := make(map[string][]*meta.Node)

	driver, err := dao.Neo4j.Create_driver()
	if err != nil {
		err := errors.Errorf("create neo4j driver failed: %s **fatal**2", err.Error()) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.Errmu.Lock()
		agentmanager.Topo.ErrCond.Wait()
		agentmanager.Topo.Errmu.Unlock()
		close(agentmanager.Topo.ErrCh)
		os.Exit(1)
	}
	defer dao.Neo4j.Close_driver(driver)

	cqlOUT = "match (n:host) return collect(distinct n.unixtime) as times"
	times, err := dao.Neo4j.General_query(cqlOUT, "times", driver)
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, err
	}

	if len(times) < 2 {
		latest = times[0]
	} else {
		latest = times[len(times)-2]
	}

	cqlOUT = fmt.Sprintf("match (nodes:`%s`) where nodes.unixtime='%s' return nodes", uuid, latest)
	single_nodes, err = dao.Neo4j.Node_query(cqlOUT, "nodes", driver)
	if err != nil {
		err = errors.Wrap(err, " **2")
		return nil, err
	}

	for _, node := range single_nodes {
		if _, ok := single_nodes_map[node.DBID]; !ok {
			single_nodes_map[node.DBID] = node
		}
	}

	for _, node := range single_nodes {
		nodes_type_map[node.Type] = append(nodes_type_map[node.Type], node)
		if node.Type == "host" {
			treerootnode = CreateTreeNode(node)
		}
	}

	for _, node := range nodes_type_map[meta.NODE_RESOURCE] {
		childnode := CreateTreeNode(node)
		treerootnode.Children = append(treerootnode.Children, childnode)
	}

	for _, node := range nodes_type_map[meta.NODE_PROCESS] {
		treenode := CreateTreeNode(node)
		treenodes_process = append(treenodes_process, treenode)
	}

	for _, node := range nodes_type_map[meta.NODE_NET] {
		treenode := CreateTreeNode(node)
		treenodes_net = append(treenodes_net, treenode)
	}

	for _, node := range treenodes_process {
		if node.Node.Metrics["Pid"] == "1" {
			node.Children = SliceToTree(treenodes_process, treenodes_net, "1")
			treerootnode.Children = append(treerootnode.Children, node)

			break
		}
	}

	return treerootnode, nil
}
