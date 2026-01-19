package messages

import (
	"net/http"
	"testing"

	ws "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/websocket"
)

func TestNullNodesEdges(t *testing.T) {
	nodesNull := []byte(`{"Nodes": null}`)

	msg := &ws.StructMessage{
		Namespace: Namespace,
		Type:      SyncMsgType,
		UUID:      "aaa",
		Status:    http.StatusOK,
		Obj:       nodesNull,
	}

	if _, _, err := UnmarshalMessage(msg); err != nil {
		t.Errorf("Should not raise error if Nodes is null: %s", err)
	}

	edgesNull := []byte(`{"Nodes": [{"ID": "aaa"}], "Edges": null}`)

	msg = &ws.StructMessage{
		Namespace: Namespace,
		Type:      SyncMsgType,
		UUID:      "bbb",
		Status:    http.StatusOK,
		Obj:       edgesNull,
	}

	if _, _, err := UnmarshalMessage(msg); err != nil {
		t.Errorf("Should not raise error if Edges is null: %s", err)
	}
}

func TestID(t *testing.T) {
	idMissing := []byte(`{"Nodes": [{"AAA": 123}]}`)

	msg := &ws.StructMessage{
		Namespace: Namespace,
		Type:      SyncMsgType,
		UUID:      "aaa",
		Status:    http.StatusOK,
		Obj:       idMissing,
	}

	if _, _, err := UnmarshalMessage(msg); err == nil {
		t.Error("Should raise an error")
	}

	idWrongType := []byte(`{"Nodes": [{"ID": 123}]}`)

	msg = &ws.StructMessage{
		Namespace: Namespace,
		Type:      SyncMsgType,
		UUID:      "aaa",
		Status:    http.StatusOK,
		Obj:       idWrongType,
	}

	if _, _, err := UnmarshalMessage(msg); err == nil {
		t.Error("Should raise an error")
	}
}

func TestHost(t *testing.T) {
	hostWrongType := []byte(`{"Nodes": [{"Host": 123}]}`)

	msg := &ws.StructMessage{
		Namespace: Namespace,
		Type:      SyncMsgType,
		UUID:      "aaa",
		Status:    http.StatusOK,
		Obj:       hostWrongType,
	}

	if _, _, err := UnmarshalMessage(msg); err == nil {
		t.Error("Should raise an error")
	}
}
