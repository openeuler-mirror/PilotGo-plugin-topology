package server

import (
	"github.com/skydive-project/skydive/api/types"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	api "github.com/skydive-project/skydive/graffiti/api/server"
	"github.com/skydive-project/skydive/graffiti/graph"
	shttp "github.com/skydive-project/skydive/graffiti/http"
)

// NodeRuleResourceHandler describes a node resource handler
type NodeRuleResourceHandler struct {
	rest.ResourceHandler
}

// NodeRuleAPI based on BasicAPIHandler
type NodeRuleAPI struct {
	rest.BasicAPIHandler
	Graph *graph.Graph
}

// Name returns resource name "noderule"
func (nrh *NodeRuleResourceHandler) Name() string {
	return "noderule"
}

// New creates a new node rule
func (nrh *NodeRuleResourceHandler) New() rest.Resource {
	return &types.NodeRule{}
}

// Update an edge rule
func (a *NodeRuleAPI) Update(id string, resource rest.Resource) (rest.Resource, bool, error) {
	return nil, false, rest.ErrNotUpdatable
}

// RegisterNodeRuleAPI register a new node rule api handler
func RegisterNodeRuleAPI(apiServer *api.Server, g *graph.Graph, authBackend shttp.AuthenticationBackend) *NodeRuleAPI {
	nra := &NodeRuleAPI{
		BasicAPIHandler: rest.BasicAPIHandler{
			ResourceHandler: &NodeRuleResourceHandler{},
			EtcdClient:      apiServer.EtcdClient,
		},
		Graph: g,
	}
	apiServer.RegisterAPIHandler(nra, authBackend)
	return nra
}
