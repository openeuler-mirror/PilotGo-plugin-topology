package server

import (
	"fmt"

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

// Create tests that resource GremlinQuery does not exists already
func (c *CaptureAPIHandler) Create(r rest.Resource, opts *rest.CreateOptions) error {
	capture := r.(*types.Capture)

	// check capabilities
	if capture.Type != "" {
		if capture.BPFFilter != "" {
			if !probes.CheckProbeCapabilities(capture.Type, probes.BPFCapability) {
				return fmt.Errorf("%s capture doesn't support BPF filtering", capture.Type)
			}
		}
		if capture.RawPacketLimit != 0 {
			if !probes.CheckProbeCapabilities(capture.Type, probes.RawPacketsCapability) {
				return fmt.Errorf("%s capture doesn't support raw packet capture", capture.Type)
			}
		}
		if capture.ExtraTCPMetric {
			if !probes.CheckProbeCapabilities(capture.Type, probes.ExtraTCPMetricCapability) {
				return fmt.Errorf("%s capture doesn't support extra TCP metrics capture", capture.Type)
			}
		}
	}

	resources := c.Index()
	for _, resource := range resources {
		resource := resource.(*types.Capture)

		sameGremlin := resource.GremlinQuery == capture.GremlinQuery
		sameBPFFilter := resource.BPFFilter == capture.BPFFilter
		sameCaptureType := resource.Type == capture.Type
		supportsMulti := capture.Type == "" || probes.CheckProbeCapabilities(capture.Type, probes.MultipleOnSameNodeCapability)

		if sameCaptureType && sameGremlin {
			if !supportsMulti || sameBPFFilter {
				return rest.ErrDuplicatedResource
			}
		}
	}

	return c.BasicAPIHandler.Create(r, opts)
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
