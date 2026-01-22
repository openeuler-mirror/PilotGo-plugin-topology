package server

import (
	"github.com/skydive-project/skydive/api/types"
	"github.com/skydive-project/skydive/flow"
	"github.com/skydive-project/skydive/flow/probes"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	api "github.com/skydive-project/skydive/graffiti/api/server"
	"github.com/skydive-project/skydive/graffiti/graph"
	shttp "github.com/skydive-project/skydive/graffiti/http"
	"github.com/skydive-project/skydive/graffiti/logging"
	ge "github.com/skydive-project/skydive/gremlin/traversal"
)

// CaptureResourceHandler describes a capture ressouce handler
type CaptureResourceHandler struct {
	rest.ResourceHandler
}

// CaptureAPIHandler based on BasicAPIHandler
type CaptureAPIHandler struct {
	rest.BasicAPIHandler
	Graph *graph.Graph
}

// Name returns "capture"
func (c *CaptureResourceHandler) Name() string {
	return "capture"
}

// New creates a new capture resource
func (c *CaptureResourceHandler) New() rest.Resource {
	return &types.Capture{
		LayerKeyMode: flow.DefaultLayerKeyModeName(),
	}
}

// Decorate populates the capture resource
func (c *CaptureAPIHandler) Decorate(resource rest.Resource) {
	capture := resource.(*types.Capture)

	count := 0

	c.Graph.RLock()
	defer c.Graph.RUnlock()

	res, err := ge.TopologyGremlinQuery(c.Graph, capture.GremlinQuery)
	if err != nil {
		logging.GetLogger().Errorf("Gremlin error: %s", err)
		return
	}

	countCaptures := func(n *graph.Node) (count int) {
		if field, err := n.GetField("Captures"); err == nil {
			if captures, ok := field.(*probes.Captures); ok {
				for _, capture := range *captures {
					if capture.State == "active" {
						count++
					}
				}

			}
		}
		return
	}

	for _, value := range res.Values() {
		switch value.(type) {
		case *graph.Node:
			n := value.(*graph.Node)
			count += countCaptures(n)
		case []*graph.Node:
			for _, n := range value.([]*graph.Node) {
				count += countCaptures(n)
			}
		default:
			count = 0
		}
	}

	capture.Count = count
}

// Update a capture
func (c *CaptureAPIHandler) Update(id string, resource rest.Resource) (rest.Resource, bool, error) {
	return nil, false, rest.ErrNotUpdatable
}

// RegisterCaptureAPI registers an new resource, capture
func RegisterCaptureAPI(apiServer *api.Server, g *graph.Graph, authBackend shttp.AuthenticationBackend) *CaptureAPIHandler {
	captureAPIHandler := &CaptureAPIHandler{
		BasicAPIHandler: rest.BasicAPIHandler{
			ResourceHandler: &CaptureResourceHandler{},
			EtcdClient:      apiServer.EtcdClient,
		},
		Graph: g,
	}
	apiServer.RegisterAPIHandler(captureAPIHandler, authBackend)
	return captureAPIHandler
}
