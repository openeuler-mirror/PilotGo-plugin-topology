package meta

import (
	"fmt"

	"github.com/pkg/errors"
)

type Nodes struct {
	Lookup map[string]*Node
	Nodes  []*Node
}

type Node struct {
	ID      string            `json:"id"` // uuid-type-basicinfo
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	UUID    string            `json:"uuid"`
	Metrics map[string]string `json:"metrics"`
}

func (ns *Nodes) Add(node *Node) {
	_, ok := ns.Lookup[node.ID]
	if ok {
		// for k, v := range node.Attrs {
		// 	if _, ok := old_n.Attrs[k]; !ok {
		// 		old_n.Attrs[k] = v
		// 	}
		// }

		return
	}

	ns.Lookup[node.ID] = node
	ns.Nodes = append(ns.Nodes, node)
}

func (ns *Nodes) Remove(id string) error {
	for i := 0; i < len(ns.Nodes); i++ {
		if ns.Nodes[i].ID != id {
			continue
		}
		ns.Nodes = append(ns.Nodes[:i], ns.Nodes[i+1:]...)
		delete(ns.Lookup, id)
		return nil
	}

	return errors.New(fmt.Sprintf("node %s not found**9", id))
}
