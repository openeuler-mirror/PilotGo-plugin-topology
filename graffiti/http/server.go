package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	auth "github.com/abbot/go-http-auth"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/graffiti/rbac"
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

func copyRequestVars(old, new *http.Request) {
	kv := gcontext.GetAll(old)
	for k, v := range kv {
		gcontext.Set(new, k, v)
	}
}

// SetTLSHeader set TLS specific headers in the response
func SetTLSHeader(w http.ResponseWriter, r *http.Request) {
	if r.TLS != nil {
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}
}

// RegisterRoutes registers a set of routes protected by an authentication backend
func (s *Server) RegisterRoutes(routes []Route, auth AuthenticationBackend) {
	for _, route := range routes {
		r := s.Router.
			Methods(route.Method).
			Name(route.Name).
			Handler(auth.Wrap(route.HandlerFunc))
		switch p := route.Path.(type) {
		case string:
			r.Path(p)
		case PathPrefix:
			r.PathPrefix(string(p))
		}
	}
}

// Listen starts listening for TCP requests
func (s *Server) Listen() error {
	listenAddrPort := fmt.Sprintf("%s:%d", s.Addr, s.Port)
	ln, err := net.Listen("tcp", listenAddrPort)
	if err != nil {
		return fmt.Errorf("Failed to listen on %s:%d: %s", s.Addr, s.Port, err)
	}

	s.listener = ln
	s.logger.Infof("Listening on socket %s:%d", s.Addr, s.Port)
	return nil
}

// Serve HTTP request
func (s *Server) Serve() {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "X-Auth-Token"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	s.Handler = handlers.CompressHandler(handlers.CORS(headersOk, originsOk, methodsOk)(s.Router))

	var err error
	if s.TLSConfig != nil {
		err = s.Server.ServeTLS(s.listener, "", "")
	} else {
		err = s.Server.Serve(s.listener)
	}

	if err == http.ErrServerClosed {
		return
	}
	s.logger.Errorf("Failed to serve on %s:%d: %s", s.Addr, s.Port, err)
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
