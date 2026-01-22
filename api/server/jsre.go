package server

import (
	"strings"

	"github.com/robertkrimen/otto"
	api "github.com/skydive-project/skydive/graffiti/api/server"
	"github.com/skydive-project/skydive/graffiti/assets"
	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/graffiti/graph/traversal"
	"github.com/skydive-project/skydive/graffiti/js"
)

// NewRuntime returns a new JavaScript runtime where accesses to resources are done
// directly using the handlers, not through HTTP requests.
func NewRuntime(g *graph.Graph, tr *traversal.GremlinTraversalParser, server *api.Server, assets assets.Assets) (*js.Runtime, error) {
	runtime, err := js.NewRuntime(assets)
	if err != nil {
		return nil, err
	}
	runtime.Start()

	queryGremlin := func(query string) otto.Value {
		ts, err := tr.Parse(strings.NewReader(query))
		if err != nil {
			return runtime.MakeCustomError("ParseError", err.Error())
		}

		result, err := ts.Exec(g, false)
		if err != nil {
			return runtime.MakeCustomError("ExecuteError", err.Error())
		}

		source, err := result.MarshalJSON()
		if err != nil {
			return runtime.MakeCustomError("MarshalError", err.Error())
		}

		r, _ := runtime.ToValue(string(source))
		return r
	}

	runtime.Set("Gremlin", func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) < 1 || !call.Argument(0).IsString() {
			return runtime.MakeCustomError("MissingQueryArgument", "Gremlin requires a string parameter")
		}

		query := call.Argument(0).String()

		return queryGremlin(query)
	})
	return runtime, nil
}
