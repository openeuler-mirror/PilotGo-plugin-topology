//go:generate sh -c "go run github.com/gomatic/renderizer --name=alert --resource=alert --type=Alert --title=Alert --article=an ../../../api/server/swagger_operations.tmpl > alert_swagger.go"
//go:generate sh -c "go run github.com/gomatic/renderizer --name=alert --resource=alert --type=Alert --title=Alert ../../../api/server/swagger_definitions.tmpl > alert_swagger.json"

package server

import (
	"time"

	"github.com/skydive-project/skydive/graffiti/api/rest"
	"github.com/skydive-project/skydive/graffiti/api/types"
	shttp "github.com/skydive-project/skydive/graffiti/http"
)

// AlertResourceHandler aims to creates and manage a new Alert.
type AlertResourceHandler struct {
	rest.ResourceHandler
}

// AlertAPIHandler aims to exposes the Alert API.
type AlertAPIHandler struct {
	rest.BasicAPIHandler
}

// New creates a new alert
func (a *AlertResourceHandler) New() rest.Resource {
	return &types.Alert{
		CreateTime: time.Now().UTC(),
	}
}

// Name returns resource name "alert"
func (a *AlertResourceHandler) Name() string {
	return "alert"
}

// RegisterAlertAPI registers an Alert's API to a designated API Server
func RegisterAlertAPI(apiServer *Server, authBackend shttp.AuthenticationBackend) *AlertAPIHandler {
	alertAPIHandler := &AlertAPIHandler{
		BasicAPIHandler: rest.BasicAPIHandler{
			ResourceHandler: &AlertResourceHandler{},
			EtcdClient:      apiServer.EtcdClient,
		},
	}
	apiServer.RegisterAPIHandler(alertAPIHandler, authBackend)
	return alertAPIHandler
}
