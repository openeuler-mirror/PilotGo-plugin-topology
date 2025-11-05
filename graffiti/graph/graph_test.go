package graph

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newGraph(t *testing.T) *Graph {
	b, err := NewMemoryBackend()
	if err != nil {
		t.Error(err.Error())
	}

	return NewGraph("testhost", b, "graph-test")
}

func TestLinks(t *testing.T) {
	g := newGraph(t)

	n1, _ := g.NewNode(GenID(), Metadata{"Value": 1, "Type": "intf"})
	n2, _ := g.NewNode(GenID(), Metadata{"Value": 2, "Type": "intf"})

	g.NewEdge(GenID(), n1, n2, nil)
	if !g.AreLinked(n1, n2, nil) {
		t.Error("nodes should be linked")
	}

	g.Unlink(n1, n2)
	if g.AreLinked(n1, n2, nil) {
		t.Error("nodes shouldn't be linked")
	}

	g.Link(n1, n2, nil)
	if !g.AreLinked(n1, n2, nil) {
		t.Error("nodes should be linked")
	}

	g.DelNode(n2)
	if g.AreLinked(n1, n2, nil) {
		t.Error("nodes shouldn't be linked")
	}
}

func TestAreLinkedWithMetadata(t *testing.T) {
	g := newGraph(t)

	n1, _ := g.NewNode(GenID(), Metadata{"Value": 1, "Type": "intf"})
	n2, _ := g.NewNode(GenID(), Metadata{"Value": 2, "Type": "intf"})

	g.Link(n1, n2, Metadata{"Type": "aaa"})
	if !g.AreLinked(n1, n2, nil) {
		t.Error("nodes should be linked")
	}

	if !g.AreLinked(n1, n2, Metadata{"Type": "aaa"}) {
		t.Error("nodes should be linked")
	}

	if g.AreLinked(n1, n2, Metadata{"Type": "bbb"}) {
		t.Error("nodes shouldn't be linked")
	}
}

func TestFirstLinkDoesExist(t *testing.T) {
	g := newGraph(t)

	n1, _ := g.NewNode(GenID(), Metadata{"Type": "intf"})
	n2, _ := g.NewNode(GenID(), Metadata{"Type": "intf"})

	g.Link(n1, n2, Metadata{})

	if g.GetFirstLink(n1, n2, nil) == nil {
		t.Error("nodes should be linked")
	}

	if g.GetFirstLink(n2, n1, nil) != nil {
		t.Error("nodes should not be linked")
	}
}

func TestFirstLinkIsCorrect(t *testing.T) {
	g := newGraph(t)

	n1, _ := g.NewNode(GenID(), Metadata{"Type": "intf"})
	n2, _ := g.NewNode(GenID(), Metadata{"Type": "intf"})

	g.Link(n2, n1, Metadata{"Field": "other"})
	expected := "me"
	g.Link(n1, n2, Metadata{"Field": expected})

	e := g.GetFirstLink(n1, n2, nil)
	if e == nil {
		t.Error("nodes should be linked")
	}

	if actual := e.Metadata["Field"]; actual != expected {
		t.Errorf("Wrong metadata['Direction'] expected '%s' got '%s'", expected, actual)
	}
}

func TestGetFirstLinkLoop(t *testing.T) {
	g := newGraph(t)

	n1, _ := g.NewNode(GenID(), Metadata{"Type": "intf"})
	n2, _ := g.NewNode(GenID(), Metadata{"Type": "intf"})

	g.Link(n2, n2, Metadata{})

	if g.GetFirstLink(n1, n2, nil) != nil {
		t.Error("nodes should not be linked")
	}

	if g.GetFirstLink(n2, n1, nil) != nil {
		t.Error("nodes should not be linked")
	}

	if g.GetFirstLink(n2, n2, nil) == nil {
		t.Error("nodes should be linked")
	}
}

func TestBasicLookup(t *testing.T) {
	g := newGraph(t)

	n1, _ := g.NewNode(GenID(), Metadata{"Value": 1, "Type": "intf"})
	n2, _ := g.NewNode(GenID(), Metadata{"Value": 2, "Type": "intf"})
	n3, _ := g.NewNode(GenID(), Metadata{"Value": 3})
	n4, _ := g.NewNode(GenID(), Metadata{"Value": 4, "Name": "Node4"})

	g.NewEdge(GenID(), n1, n2, nil)
	g.NewEdge(GenID(), n2, n3, nil)
	g.Link(n1, n4, nil)

	if n1.ID != g.GetNode(n1.ID).ID {
		t.Error("Wrong node returned")
	}

	if n1.ID != g.LookupFirstNode(Metadata{"Value": 1}).ID {
		t.Error("Wrong node returned")
	}

	r := g.GetNodes(Metadata{"Type": "intf"})
	if len(r) != 2 {
		t.Error("Wrong number of nodes returned")
	}

	for i := range r {
		if !(r[i].ID == n1.ID || r[i].ID == n2.ID) {
			t.Error("Wrong nodes returned")
		}
	}
}

func TestBasicLookupUnsupportedTypes(t *testing.T) {
	g := newGraph(t)

	n1, _ := g.NewNode(GenID(), Metadata{"Value": int32(1), "Type": float64(44.5)})

	if n := g.LookupFirstNode(Metadata{"Value": uint64(1)}); n == nil || n.ID != n1.ID {
		t.Error("Wrong node returned")
	}

	if n := g.LookupFirstNode(Metadata{"Value": 1}); n == nil || n.ID != n1.ID {
		t.Error("Wrong node returned")
	}

	if n1.ID != g.LookupFirstNode(Metadata{"Type": 44.5}).ID {
		t.Error("Wrong node returned")
	}
}

func TestHierarchyLookup(t *testing.T) {
	g := newGraph(t)

	n1, _ := g.NewNode(GenID(), Metadata{"Value": 1, "Type": "intf"})
	n2, _ := g.NewNode(GenID(), Metadata{"Value": 2, "Type": "intf"})
	n3, _ := g.NewNode(GenID(), Metadata{"Value": 3})
	n4, _ := g.NewNode(GenID(), Metadata{"Value": 4, "Name": "Node4"})

	g.Link(n1, n2, nil)
	g.Link(n2, n3, nil)
	g.Link(n3, n4, nil)
	g.Link(n2, n4, nil)

	r := g.LookupParents(n4, nil, nil)
	if len(r) != 2 {
		t.Error("Wrong number of nodes returned")
	}

	for i := range r {
		if !(r[i].ID == n2.ID || r[i].ID == n3.ID) {
			t.Error("Wrong nodes returned")
		}
	}

	r = g.LookupParents(n4, Metadata{"Type": "intf"}, nil)
	if len(r) != 1 {
		t.Error("Wrong number of nodes returned")
	}
	if r[0].ID != n2.ID {
		t.Error("Wrong nodes returned")
	}

	r = g.LookupChildren(n2, nil, nil)
	if len(r) != 2 {
		t.Errorf("Wrong number of nodes returned: %+v", r)
	}

	for i := range r {
		if !(r[i].ID == n3.ID || r[i].ID == n4.ID) {
			t.Error("Wrong nodes returned")
		}
	}

	r = g.LookupChildren(n2, Metadata{"Name": "Node4"}, nil)
	if len(r) != 1 {
		t.Error("Wrong number of nodes returned")
	}

	if r[0].ID != n4.ID {
		t.Error("Wrong nodes returned")
	}
}

func TestPath(t *testing.T) {
	g := newGraph(t)

	validatePath := func(nodes []*Node, expected string) bool {
		var values []string

		for _, n := range nodes {
			value, _ := n.GetFieldInt64("Value")
			values = append(values, strconv.FormatInt(value, 10))
		}

		return expected == strings.Join(values, "/")
	}

	n1, _ := g.NewNode(GenID(), Metadata{"Value": 1, "Type": "intf"})
	n2, _ := g.NewNode(GenID(), Metadata{"Value": 2, "Type": "intf"})
	n3, _ := g.NewNode(GenID(), Metadata{"Value": 3})
	n4, _ := g.NewNode(GenID(), Metadata{"Value": 4, "Name": "Node4"})

	g.Link(n1, n2, Metadata{"Type": "Layer2"})
	g.Link(n2, n3, Metadata{"Type": "Layer2"})
	g.Link(n3, n4, Metadata{"Type": "Layer2"})

	r := g.LookupShortestPath(n4, Metadata{"Value": 1}, nil)
	if len(r) == 0 || !validatePath(r, "4/3/2/1") {
		t.Errorf("Wrong nodes returned: %v", r)
	}

	// add a shorter link
	g.Link(n4, n1, Metadata{"Type": "Layer2"})
	r = g.LookupShortestPath(n4, Metadata{"Value": 1}, nil)
	if len(r) == 0 || !validatePath(r, "4/1") {
		t.Errorf("Wrong nodes returned: %v", r)
	}
	g.Unlink(n4, n1)

	r = g.LookupShortestPath(n4, Metadata{"Value": 2}, nil)
	if len(r) == 0 || !validatePath(r, "4/3/2") {
		t.Errorf("Wrong nodes returned: %v", r)
	}

	r = g.LookupShortestPath(n4, Metadata{"Value": 55}, nil)
	if len(r) > 0 {
		t.Errorf("Shouldn't have true returned: %v", r)
	}

	r = g.LookupShortestPath(n4, Metadata{"Value": 2}, Metadata{"Type": "Layer6"})
	if len(r) > 0 {
		t.Errorf("Shouldn't have true returned: %v", r)
	}

	// add a shorter link in order to validate edge validator
	g.Link(n1, n4, Metadata{"Type": "Layer3"})

	r = g.LookupShortestPath(n4, Metadata{"Value": 1}, Metadata{"Type": "Layer2"})
	if len(r) == 0 || !validatePath(r, "4/3/2/1") {
		t.Errorf("Wrong nodes returned: %v", r)
	}

	r = g.LookupShortestPath(n4, Metadata{"Value": 1}, Metadata{"Type": "Layer3"})
	if len(r) == 0 || !validatePath(r, "4/1") {
		t.Errorf("Wrong nodes returned: %v", r)
	}
	g.Unlink(n1, n4)

	// test shortestPath on the following graph
	// n1 -- n2 -- n3 -- n4 -----------
	//  \                              \
	//   \-- n11 -- n12---------------- n5
	//                \                /
	//                 \-- n121 -- n122
	n5, _ := g.NewNode(GenID(), Metadata{"Value": 5, "Name": "Node5"})
	g.Link(n4, n5, Metadata{"Type": "Layer2"})

	n11, _ := g.NewNode(GenID(), Metadata{"Value": 11, "Name": "Node11"})
	n12, _ := g.NewNode(GenID(), Metadata{"Value": 12, "Name": "Node12"})
	g.Link(n1, n11, Metadata{"Type": "Layer2"})
	g.Link(n11, n12, Metadata{"Type": "Layer2"})
	g.Link(n12, n5, Metadata{"Type": "Layer2"})

	n121, _ := g.NewNode(GenID(), Metadata{"Value": 121, "Name": "Node121"})
	n122, _ := g.NewNode(GenID(), Metadata{"Value": 122, "Name": "Node122"})

	g.Link(n12, n121, Metadata{"Type": "Layer2"})
	g.Link(n121, n122, Metadata{"Type": "Layer2"})
	g.Link(n122, n5, Metadata{"Type": "Layer2"})

	r = g.LookupShortestPath(n1, Metadata{"Value": 5}, nil)
	if len(r) == 0 || !validatePath(r, "1/11/12/5") {
		t.Errorf("Wrong nodes returned: %v", r)
	}
}

