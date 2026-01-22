package server

import (
	"github.com/skydive-project/skydive/config"
	shttp "github.com/skydive-project/skydive/graffiti/http"
)

type configAPI struct {
	cfg *config.SkydiveConfig
}

func (c *configAPI) registerEndpoints(r *shttp.Server, authBackend shttp.AuthenticationBackend) {
	// swagger:operation GET /config/{key} getConfig
	//
	// Get configuration value
	//
	// ---
	// summary: Get configuration
	//
	// tags:
	// - Config
	//
	// produces:
	// - application/json
	//
	// schemes:
	// - http
	// - https
	//
	// parameters:
	// - name: key
	//   in: path
	//   required: true
	//   type: string
	//
	// responses:
	//   200:
	//     description: configuration value
	//     schema:
	//       $ref: '#/definitions/AnyValue'

	routes := []shttp.Route{
		{
			Name:        "ConfigGet",
			Method:      "GET",
			Path:        "/api/config/{key}",
			HandlerFunc: c.configGet,
		},
	}

	r.RegisterRoutes(routes, authBackend)
}

// RegisterConfigAPI registers a configuration endpoint (read only) in API server
func RegisterConfigAPI(r *shttp.Server, authBackend shttp.AuthenticationBackend) {
	c := &configAPI{
		cfg: config.GetConfig(),
	}

	c.registerEndpoints(r, authBackend)
}
