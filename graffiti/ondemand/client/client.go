package client

import (
	"encoding/json"
	"net/http"

	cache "github.com/pmylund/go-cache"
	"github.com/safchain/insanelock"
	"github.com/skydive-project/go-debouncer"

	"github.com/skydive-project/skydive/graffiti/api/rest"
	etcd "github.com/skydive-project/skydive/graffiti/etcd/client"
	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/graffiti/ondemand"
	ws "github.com/skydive-project/skydive/graffiti/websocket"
)

// OnDemandNodeResource describes an association between a graph
// node and its corresponding on-demand resource
type OnDemandNodeResource struct {
	Node     *graph.Node
	Resource rest.Resource
}

// OnDemandClientHandler is the interface to be implemented by ondemand clients
type OnDemandClientHandler interface {
	ResourceName() string
	GetNodeResources(resource rest.Resource) []OnDemandNodeResource
	CheckState(n *graph.Node, resource rest.Resource) bool
	DecodeMessage(msg json.RawMessage) (rest.Resource, error)
	EncodeMessage(nodeID graph.Identifier, resource rest.Resource) (json.RawMessage, error)
}

// OnDemandClient describes an ondemand task client based on a websocket
type OnDemandClient struct {
	insanelock.RWMutex
	etcd.MasterElection
	graph.DefaultGraphListener
	graph                   *graph.Graph
	apiHandler              rest.WatchableHandler
	agentPool               ws.StructSpeakerPool
	subscriberPool          ws.StructSpeakerPool
	wsNamespace             string
	wsNotificationNamespace string
	resources               map[string]rest.Resource
	watcher                 rest.StoppableWatcher
	registeredNodes         map[graph.Identifier]map[string]bool
	deletedNodeCache        *cache.Cache
	checkForRegistration    *debouncer.Debouncer
	resourceName            string
	handler                 OnDemandClientHandler
}

type handlerNodeState struct {
	uuid    string
	started bool
}

type nodeTask struct {
	id       graph.Identifier
	host     string
	resource rest.Resource
}

func (o *OnDemandClient) removeRegisteredNode(nodeID graph.Identifier, resourceID string) {
	o.Lock()
	tasks, found := o.registeredNodes[nodeID]
	if !found {
		o.Unlock()
		return
	}
	delete(tasks, resourceID)
	if len(o.registeredNodes) == 0 {
		delete(o.registeredNodes, nodeID)
	}
	o.Unlock()
}

// OnStructMessage event, valid message type: StartReply or StopReply message
func (o *OnDemandClient) OnStructMessage(c ws.Speaker, m *ws.StructMessage) {
	var rawQuery ondemand.RawQuery
	if err := json.Unmarshal(m.Obj, &rawQuery); err != nil {
		logging.GetLogger().Errorf("unable to decode %s message: %s", o.resourceName, err)
		return
	}

	resource, err := o.handler.DecodeMessage(rawQuery.Resource)
	if err != nil {
		logging.GetLogger().Errorf("unable to decode %s: %s", o.resourceName, err)
		return
	}

	switch m.Type {
	case "StartReply":
		// not registered thus remove from registered cache
		if m.Status != http.StatusOK {
			logging.GetLogger().Debugf("%s start request failed %v", o.resourceName, m.Debug())
			o.removeRegisteredNode(rawQuery.NodeID, resource.GetID())
		} else {
			logging.GetLogger().Debugf("%s start request succeeded %v", o.resourceName, m.Debug())
		}
		o.subscriberPool.BroadcastMessage(ws.NewStructMessage(o.wsNotificationNamespace, "NodeUpdated", resource))
	case "StopReply":
		if m.Status == http.StatusOK {
			o.removeRegisteredNode(rawQuery.NodeID, resource.GetID())
		} else {
			logging.GetLogger().Debugf("%s stop request failed %v", o.resourceName, m.Debug())
		}
	}
}

func (o *OnDemandClient) registerTasks(nps map[graph.Identifier]nodeTask) {
	for _, np := range nps {
		o.registerTask(np)
	}
}

func (o *OnDemandClient) registerTask(np nodeTask) bool {
	body, err := o.handler.EncodeMessage(np.id, np.resource)
	if err != nil {
		logging.GetLogger().Errorf("Unable to encode message for agent %s: %s", np.host, err)
		return false
	}

	msg := ws.NewStructMessage(o.wsNamespace, "Start", ondemand.RawQuery{
		NodeID:   np.id,
		Resource: body,
	})

	if err := o.agentPool.SendMessageTo(msg, np.host); err != nil {
		logging.GetLogger().Errorf("Unable to send message to agent %s: %s", np.host, err)
		return false
	}

	o.Lock()
	if _, found := o.registeredNodes[np.id]; !found {
		o.registeredNodes[np.id] = make(map[string]bool)
	}
	o.registeredNodes[np.id][np.resource.GetID()] = false
	o.Unlock()

	logging.GetLogger().Debugf("Registered task on %s with resource %s", np.id, np.resource.GetID())

	return true
}

func (o *OnDemandClient) unregisterTask(node *graph.Node, resource rest.Resource) bool {
	body, err := o.handler.EncodeMessage(node.ID, resource)
	if err != nil {
		logging.GetLogger().Errorf("Unable to encode message for agent %s: %s", node.Host, err)
		return false
	}

	msg := ws.NewStructMessage(o.wsNamespace, "Stop", ondemand.RawQuery{
		NodeID:   node.ID,
		Resource: body,
	})

	if err := o.agentPool.SendMessageTo(msg, node.Host); err != nil {
		logging.GetLogger().Errorf("Unable to send message to agent %s: %s", node.Host, err)
		return false
	}

	o.removeRegisteredNode(node.ID, resource.GetID())

	return true
}

func (o *OnDemandClient) nodeTasks(nrs []OnDemandNodeResource) map[graph.Identifier]nodeTask {
	toRegister := func(nr OnDemandNodeResource) (nodeID graph.Identifier, host string, register bool) {
		// check not already registered
		tasks, ok := o.registeredNodes[nr.Node.ID]
		if ok {
			ok = tasks[nr.Resource.GetID()]
		}

		if ok {
			logging.GetLogger().Debugf("%s already registered on %s", nr.Resource.GetID(), nr.Node.ID)
			return
		}

		return nr.Node.ID, nr.Node.Host, true
	}

	nps := map[graph.Identifier]nodeTask{}
	for _, nr := range nrs {
		if nodeID, host, ok := toRegister(nr); ok {
			nps[nodeID] = nodeTask{nodeID, host, nr.Resource}
		}
	}

	return nps
}
