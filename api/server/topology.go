package server

import (
	"errors"
	"io"

	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/graffiti/graph/traversal"
)

func shortID(s graph.Identifier) graph.Identifier {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// dotMarshaller outputs a GraversalStep as dot
func dotMarshaller(step traversal.GraphTraversalStep, w io.Writer) error {
	if graphTraversal, ok := step.(*traversal.GraphTraversal); ok {
		graphToDot(w, graphTraversal.Graph)
		return nil
	}
	return errors.New("Only graph can be outputted as dot")
}
