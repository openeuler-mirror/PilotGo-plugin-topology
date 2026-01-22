package server

import (
	"github.com/skydive-project/skydive/graffiti/api/rest"
	"github.com/skydive-project/skydive/graffiti/graph"
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
