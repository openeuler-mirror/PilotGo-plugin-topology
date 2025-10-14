package http

import (
	"net/http"

	"github.com/abbot/go-http-auth"
	"github.com/gorilla/context"
)

// NoAuthenticationBackend describes an authenticate backed that
// allows everyone to do anything
type NoAuthenticationBackend struct {
}

// Name returns the name of the backend
func (n *NoAuthenticationBackend) Name() string {
	return "noauth"
}

// DefaultUserRole returns the name of the backend
func (n *NoAuthenticationBackend) DefaultUserRole(user string) string {
	return DefaultUserRole
}

// SetDefaultUserRole defines the default user role
func (n *NoAuthenticationBackend) SetDefaultUserRole(role string) {
}

// Authenticate the user and its password
func (n *NoAuthenticationBackend) Authenticate(username string, password string) (string, error) {
	return "", nil
}

// Wrap an HTTP handler with no authentication backend
func (n *NoAuthenticationBackend) Wrap(wrapped auth.AuthenticatedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ar := &auth.AuthenticatedRequest{Request: *r, Username: "admin"}
		copyRequestVars(r, &ar.Request)
		wrapped(w, ar)
		context.Clear(&ar.Request)
	}
}

// NewNoAuthenticationBackend returns a new authentication backend that allows
// everyone to do anything
func NewNoAuthenticationBackend() *NoAuthenticationBackend {
	return &NoAuthenticationBackend{}
}

// NoAuthenticationWrap wraps a handler with no authentication
func NoAuthenticationWrap(wrapped auth.AuthenticatedHandlerFunc) http.HandlerFunc {
	return NewNoAuthenticationBackend().Wrap(wrapped)
}
