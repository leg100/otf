package html

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/html/paths"
)

// UIRouter wraps the given router with a router suitable for web UI routes.
func UIRouter(r *mux.Router) *mux.Router {
	r = r.PathPrefix(paths.UIPrefix).Subrouter()
	return r
}
