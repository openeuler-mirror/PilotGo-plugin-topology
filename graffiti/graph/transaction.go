package graph

import (
	"reflect"
)

// MetadataTransaction describes a metadata transaction in the graph
type MetadataTransaction struct {
	graph        *Graph
	graphElement interface{}
	adds         map[string]interface{}
	removes      []string
}

// AddMetadata in the current transaction
func (t *MetadataTransaction) AddMetadata(k string, v interface{}) {
	t.adds[k] = v
}

// DelMetadata in the current transaction
func (t *MetadataTransaction) DelMetadata(k string) {
	t.removes = append(t.removes, k)
}

// Commit the current transaction to the graph
func (t *MetadataTransaction) Commit() error {
	var e *graphElement
	var kind graphEventType

	switch t.graphElement.(type) {
	case *Node:
		e = &t.graphElement.(*Node).graphElement
		kind = NodeUpdated
	case *Edge:
		e = &t.graphElement.(*Edge).graphElement
		kind = EdgeUpdated
	}

	var ops []PartiallyUpdatedOp
	var updated bool
	for k, v := range t.adds {
		if o, ok := e.Metadata[k]; ok && reflect.DeepEqual(o, v) {
			continue
		}

		if e.Metadata.SetField(k, v) {
			updated = true
			ops = append(ops, PartiallyUpdatedOp{
				Type:  PartiallyUpdatedAddOpType,
				Key:   k,
				Value: v,
			})
		}
	}

	for _, k := range t.removes {
		if e.Metadata.DelField(k) {
			ops = append(ops, PartiallyUpdatedOp{
				Type: PartiallyUpdatedDelOpType,
				Key:  k,
			})
			updated = true
		}
	}
	if !updated {
		return nil
	}

	e.UpdatedAt = TimeUTC()
	e.Revision++

	if err := t.graph.backend.MetadataUpdated(t.graphElement); err != nil {
		return err
	}

	t.graph.eventHandler.NotifyEvent(kind, t.graphElement, ops...)

	return nil
}
