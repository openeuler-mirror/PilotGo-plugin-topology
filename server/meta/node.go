package meta

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type Nodes struct {
	Lock         sync.Mutex
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

		for j, n := range ns.LookupByType[ns.Nodes[i].Type] {
			if n.ID == node.ID {
				ns.LookupByType[ns.Nodes[i].Type] = append(ns.LookupByType[ns.Nodes[i].Type][:j], ns.LookupByType[ns.Nodes[i].Type][j+1:]...)
				break
			}
		}

		for j, n := range ns.LookupByUUID[ns.Nodes[i].UUID] {
			if n.ID == node.ID {
				ns.LookupByUUID[ns.Nodes[i].Type] = append(ns.LookupByUUID[ns.Nodes[i].UUID][:j], ns.LookupByUUID[ns.Nodes[i].UUID][j+1:]...)
				break
			}
		}

		ns.Nodes = append(ns.Nodes[:i], ns.Nodes[i+1:]...)
		delete(ns.Lookup, node.ID)

		return nil
	}

	return errors.New(fmt.Sprintf("node %s not found**9", node.ID))
}
