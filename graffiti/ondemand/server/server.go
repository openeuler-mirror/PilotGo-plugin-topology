package server

import (
	"encoding/json"
	"errors"

	"github.com/safchain/insanelock"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/graffiti/ondemand"
	ws "github.com/skydive-project/skydive/graffiti/websocket"
)

type activeTask struct {
	graph    *graph.Graph
	node     *graph.Node
	resource rest.Resource
	task     ondemand.Task
	handler  OnDemandServerHandler
}

// OnDemandServer describes an ondemand task server based on websocket
type OnDemandServer struct {
	insanelock.RWMutex
	graph.DefaultGraphListener
	ws.DefaultSpeakerEventHandler
	Graph        *graph.Graph
	clientPool   *ws.StructClientPool
	activeTasks  map[graph.Identifier]map[string]*activeTask
	wsNamespace  string
	resourceName string
	handler      OnDemandServerHandler
}

// OnDemandServerHandler is the interface to be implemented by ondemand servers
type OnDemandServerHandler interface {
	ResourceName() string
	DecodeMessage(msg json.RawMessage) (rest.Resource, error)
	CreateTask(*graph.Node, rest.Resource) (interface{}, error)
	RemoveTask(*graph.Node, rest.Resource, interface{}) error
}

// ErrTaskNotFound used when a task is not found for a specific node
var ErrTaskNotFound = errors.New("task not found")

func (o *OnDemandServer) registerTask(n *graph.Node, resource rest.Resource) bool {
	logging.GetLogger().Debugf("Attempting to register %s %s on node %s", o.resourceName, resource.GetID(), n.ID)

	if _, err := n.GetFieldString("Type"); err != nil {
		logging.GetLogger().Infof("Unable to register task type of node unknown %v", n)
		return false
	}

	o.Lock()
	defer o.Unlock()

	if tasks, active := o.activeTasks[n.ID]; active {
		if _, found := tasks[resource.GetID()]; found {
			logging.GetLogger().Debugf("A task already exists for %s on node %s", resource.GetID(), n.ID)
			return false
		}
	}

	task, err := o.handler.CreateTask(n, resource)
	if err != nil {
		logging.GetLogger().Errorf("Failed to register %s task: %s", o.resourceName, err)
		return false
	}

	active := &activeTask{
		graph:    o.Graph,
		node:     n,
		resource: resource,
		task:     task,
		handler:  o.handler,
	}

	if _, found := o.activeTasks[n.ID]; !found {
		o.activeTasks[n.ID] = make(map[string]*activeTask)
	}
	o.activeTasks[n.ID][resource.GetID()] = active

	logging.GetLogger().Debugf("New active task on: %v (%v)", n, resource)
	return true
}
