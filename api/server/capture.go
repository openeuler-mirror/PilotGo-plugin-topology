package server

import (
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"github.com/skydive-project/skydive/api/types"
	"github.com/skydive-project/skydive/flow"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	shttp "github.com/skydive-project/skydive/graffiti/http"
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
