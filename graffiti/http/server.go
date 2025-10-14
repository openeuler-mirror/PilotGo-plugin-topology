package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"

	auth "github.com/abbot/go-http-auth"
	"github.com/gorilla/mux"

	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/graffiti/service"
)

// PathPrefix describes the prefix of the path of an URL
type PathPrefix string

// Route describes an HTTP route with a name, a HTTP verb,
// a path protected by an authentication backend
type Route struct {
	Name        string
	Method      string
	Path        interface{}
	HandlerFunc auth.AuthenticatedHandlerFunc
}

// Server describes a HTTP server for a service that dispatches requests to routes
type Server struct {
	sync.RWMutex
	http.Server
	Host        string
	ServiceType service.Type
	Router      *mux.Router
	Addr        string
	Port        int
	lock        sync.Mutex
	listener    net.Listener
	wg          sync.WaitGroup
	logger      logging.Logger
}

// NewServer returns a new HTTP service for a service
func NewServer(host string, serviceType service.Type, addr string, port int, tlsConfig *tls.Config, logger logging.Logger) *Server {
	if logger == nil {
		logger = logging.GetLogger()
	}

	router := mux.NewRouter().StrictSlash(true)
	router.Headers("X-Host-ID", host, "X-Service-Type", serviceType.String())

	return &Server{
		Server: http.Server{
			TLSConfig: tlsConfig,
		},
		Host:        host,
		ServiceType: serviceType,
		Router:      router,
		Addr:        addr,
		Port:        port,
		logger:      logger,
	}
}
