package server

import (
	"github.com/skydive-project/skydive/api/types"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	api "github.com/skydive-project/skydive/graffiti/api/server"
	"github.com/skydive-project/skydive/graffiti/graph"
	shttp "github.com/skydive-project/skydive/graffiti/http"
)

// EdgeRuleResourceHandler describes a edge rule resource handler
type EdgeRuleResourceHandler struct {
	rest.ResourceHandler
}

// EdgeRuleAPI based on BasicAPIHandler
type EdgeRuleAPI struct {
	rest.BasicAPIHandler
	Graph *graph.Graph
}

// Name returns resource name "edgerule"
func (erh *EdgeRuleResourceHandler) Name() string {
	return "edgerule"
}

// New creates a new EdgeRule
func (erh *EdgeRuleResourceHandler) New() rest.Resource {
	return &types.EdgeRule{}
}

// Update an edge rule
func (a *EdgeRuleAPI) Update(id string, resource rest.Resource) (rest.Resource, bool, error) {
	return nil, false, rest.ErrNotUpdatable
}

// RegisterEdgeRuleAPI registers an EdgeRule's API to a designated API Server
func RegisterEdgeRuleAPI(apiServer *api.Server, g *graph.Graph, authBackend shttp.AuthenticationBackend) *EdgeRuleAPI {
	era := &EdgeRuleAPI{
		BasicAPIHandler: rest.BasicAPIHandler{
			ResourceHandler: &EdgeRuleResourceHandler{},
			EtcdClient:      apiServer.EtcdClient,
		},
		Graph: g,
	}
	apiServer.RegisterAPIHandler(era, authBackend)
	return era
}
