package graph

import (
	"sync"

	"github.com/pkg/errors"
)

type Nodes struct {
	Lock         sync.Mutex

	// key: node.id
	Lookup       map[string]*Node
	
	LookupByType map[string][]*Node
	LookupByUUID map[string][]*Node
	Nodes        []*Node
}

type Node struct {
	DBID       int64             `json:"dbid"`
	ID         string            `json:"id"` // uuid-type-basicinfo
	Name       string            `json:"name"`
	Type       string            `json:"Type"`
	UUID       string            `json:"uuid"`
	Unixtime   string            `json:"unixtime"`
	Tags       []string          `json:"tags"`
	LayoutAttr string            `json:"layoutattr"`
	ComboId    string            `json:"comboId"`
	Network    []Netconnection   `json:"network"`
	Metrics    map[string]string `json:"metrics"`
}

func (ns *Nodes) Add(node *Node) {
	ns.Lock.Lock()
	defer ns.Lock.Unlock()
	if _, ok := ns.Lookup[node.ID]; !ok {
		ns.Lookup[node.ID] = node
		ns.LookupByType[node.Type] = append(ns.LookupByType[node.Type], node)
		ns.LookupByUUID[node.UUID] = append(ns.LookupByUUID[node.UUID], node)
		ns.Nodes = append(ns.Nodes, node)
	}
}

func (ns *Nodes) Remove(node *Node) error {
	for i := 0; i < len(ns.Nodes); i++ {
		if ns.Nodes[i].ID != node.ID {
			continue
		}
		// 移除ns.lookupbytype中的node节点
		if typenodes, ok := ns.LookupByType[ns.Nodes[i].Type]; ok && len(typenodes) > 0 {
			for j, n := range typenodes {
				if n.ID == node.ID {
					ns.LookupByType[ns.Nodes[i].Type] = append(typenodes[:j], typenodes[j+1:]...)
					break
				}
			}
		} else {
			return errors.Errorf("failed to remove node: %v from nodes.lookupbytype", node)
		}
		// 移除ns.lookupbyuuid中的node节点
		if uuidnodes, ok := ns.LookupByUUID[ns.Nodes[i].UUID]; ok && len(uuidnodes) > 0 {
			for j, n := range uuidnodes {
				if n.ID == node.ID {
					ns.LookupByUUID[ns.Nodes[i].UUID] = append(uuidnodes[:j], uuidnodes[j+1:]...)
					break
				}
			}
		} else {
			return errors.Errorf("failed to remove node: %v from nodes.lookupbyuuid", node)
		}
		// 移除ns.nodes和ns.lookup中的node节点
		ns.Nodes = append(ns.Nodes[:i], ns.Nodes[i+1:]...)
		delete(ns.Lookup, node.ID)

		return nil
	}

	return errors.Errorf("node %s not found", node.ID)
}
