package graph

import (
	"testing"
)

func TestAddEdgeMissingNode(t *testing.T) {
	b, err := NewMemoryBackend()
	if err != nil {
		t.Error(err)
	}

	e := &Edge{
		Parent: "aaa",
		Child:  "bbb",
		graphElement: graphElement{
			ID:   GenID(),
			Host: "host",
		},
	}

	if err := b.EdgeAdded(e); err == nil {
		t.Errorf("Edge inserted with missing nodes: %s", err)
	}
}
