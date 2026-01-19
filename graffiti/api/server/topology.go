package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	auth "github.com/abbot/go-http-auth"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/api/types"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph/traversal"
	shttp "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/http"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/rbac"
)

// TopologyMarshaller is used to output a gremlin step
// into a specific format
type TopologyMarshaller func(traversal.GraphTraversalStep, io.Writer) error

// TopologyMarshallers maps Accept headers to topology marshallers
type TopologyMarshallers map[string]TopologyMarshaller

// TopologyAPI exposes the topology query API
type TopologyAPI struct {
	graph            *graph.Graph
	gremlinParser    *traversal.GremlinTraversalParser
	extraMarshallers TopologyMarshallers
}

func (t *TopologyAPI) topologyIndex(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	if !rbac.Enforce(r.Username, "topology", "read") {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// use a buffer to render the result in order to limit the lock time
	// if the client is slow
	var b bytes.Buffer

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	t.graph.RLock()
	if err := json.NewEncoder(&b).Encode(t.graph); err != nil {
		t.graph.RUnlock()
		http.Error(w, fmt.Sprintf("Error while encoding response: %s", err), http.StatusNotAcceptable)
		return
	}
	logging.GetLogger().Info("Graph: %s", b.String())

	t.graph.RUnlock()

	if _, err := w.Write(b.Bytes()); err != nil {
		logging.GetLogger().Errorf("Error while writing response: %s", err)
	}
}

func (t *TopologyAPI) topologySearch(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	if !rbac.Enforce(r.Username, "topology", "read") {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	resource := types.TopologyParams{}
	data, _ := ioutil.ReadAll(r.Body)
	if len(data) != 0 {
		if err := json.Unmarshal(data, &resource); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if resource.GremlinQuery == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ts, err := t.gremlinParser.Parse(strings.NewReader(resource.GremlinQuery))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := ts.Exec(t.graph, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// use a buffer to render the result in order to limit the lock time
	// if the client is slow
	var b bytes.Buffer

	acceptHeader := r.Header.Get("Accept")
	if marshaller, found := t.extraMarshallers[acceptHeader]; found {
		err = marshaller(res, &b)
	} else {
		err = json.NewEncoder(&b).Encode(res)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error while encoding response: %s", err), http.StatusNotAcceptable)
		return
	}

	logging.GetLogger().Info("Query result: %s", b.String())

	w.Header().Set("Content-Type", acceptHeader+"; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(b.Bytes()); err != nil {
		logging.GetLogger().Errorf("Error while writing response: %s", err)
	}
}

func (t *TopologyAPI) registerEndpoints(r *shttp.Server, authBackend shttp.AuthenticationBackend) {
	// swagger:operation GET /topology getTopology
	//
	// Get topology
	//
	// ---
	// summary: Get topology
	//
	// tags:
	// - topology
	//
	// consumes:
	// - application/json
	//
	// produces:
	// - application/json
	//
	// schemes:
	// - http
	// - https
	//
	// responses:
	//   200:
	//     description: topology
	//     schema:
	//       type: object

	// swagger:operation POST /topology searchTopology
	//
	// Search topology
	//
	// ---
	// summary: Search topology
	//
	// tags:
	// - topology
	//
	// consumes:
	// - application/json
	//
	// produces:
	// - application/json
	//
	// schemes:
	// - http
	// - https
	//
	// parameters:
	//   - in: body
	//     name: params
	//     required: true
	//     schema:
	//       $ref: '#/definitions/TopologyParams'
	//
	// responses:
	//   200:
	//     description: query result
	//     schema:
	//       $ref: '#/definitions/AnyValue'
	//   204:
	//     description: empty query

	routes := []shttp.Route{
		{
			Name:        "TopologiesIndex",
			Method:      "GET",
			Path:        "/api/topology",
			HandlerFunc: t.topologyIndex,
		},
		{
			Name:        "TopologiesSearch",
			Method:      "POST",
			Path:        "/api/topology",
			HandlerFunc: t.topologySearch,
		},
	}

	r.RegisterRoutes(routes, authBackend)
}

// RegisterTopologyAPI registers a new topology query API
func RegisterTopologyAPI(r *shttp.Server, g *graph.Graph, parser *traversal.GremlinTraversalParser, authBackend shttp.AuthenticationBackend, extraMarshallers map[string]TopologyMarshaller) {
	t := &TopologyAPI{
		gremlinParser:    parser,
		graph:            g,
		extraMarshallers: extraMarshallers,
	}

	t.registerEndpoints(r, authBackend)
}
