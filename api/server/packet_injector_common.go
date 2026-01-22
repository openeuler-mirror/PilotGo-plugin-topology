package server

import (
	"github.com/skydive-project/skydive/api/types"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	api "github.com/skydive-project/skydive/graffiti/api/server"
	"github.com/skydive-project/skydive/graffiti/graph"
	shttp "github.com/skydive-project/skydive/graffiti/http"
	ge "github.com/skydive-project/skydive/gremlin/traversal"
)

type packetInjectorResourceHandler struct {
	rest.ResourceHandler
}

// PacketInjectorAPI exposes the packet injector API
type PacketInjectorAPI struct {
	rest.BasicAPIHandler
	Graph *graph.Graph
}

func (pirh *packetInjectorResourceHandler) Name() string {
	return "injectpacket"
}

func (pirh *packetInjectorResourceHandler) New() rest.Resource {
	return &types.PacketInjection{}
}

// Create allocates a new packet injection
func (pi *PacketInjectorAPI) Create(r rest.Resource, opts *rest.CreateOptions) error {
	ppr := r.(*types.PacketInjection)

	if err := pi.validateRequest(ppr); err != nil {
		return err
	}
	e := pi.BasicAPIHandler.Create(ppr, opts)
	return e
}

// Update a packet injection
func (pi *PacketInjectorAPI) Update(id string, resource rest.Resource) (rest.Resource, bool, error) {
	return nil, false, rest.ErrNotUpdatable
}

func (pi *PacketInjectorAPI) getNode(gremlinQuery string) *graph.Node {
	res, err := ge.TopologyGremlinQuery(pi.Graph, gremlinQuery)
	if err != nil {
		return nil
	}

	for _, value := range res.Values() {
		switch value.(type) {
		case *graph.Node:
			return value.(*graph.Node)
		default:
			return nil
		}
	}
	return nil
}

// RegisterPacketInjectorAPI registers a new packet injector resource in the API
func RegisterPacketInjectorAPI(g *graph.Graph, apiServer *api.Server, authBackend shttp.AuthenticationBackend) *PacketInjectorAPI {
	pia := &PacketInjectorAPI{
		BasicAPIHandler: rest.BasicAPIHandler{
			ResourceHandler: &packetInjectorResourceHandler{},
			EtcdClient:      apiServer.EtcdClient,
		},
		Graph: g,
	}
	apiServer.RegisterAPIHandler(pia, authBackend)
	return pia
}
