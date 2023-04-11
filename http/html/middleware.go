package html

import (
	"github.com/gorilla/mux"
)

// UIRouter wraps the given router with a router suitable for web UI routes.
func UIRouter(r *mux.Router) *mux.Router {
	r = r.PathPrefix("/app").Subrouter()
	return r
}
