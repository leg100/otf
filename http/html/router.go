package html

import (
	"net/http"

	"github.com/gorilla/mux"
)

// router wraps mux's router, adding various helper methods
type router struct {
	*mux.Router
}

// sub turns mux into chi (almost)
func (r *router) sub(group func(r *router)) {
	group(&router{r.NewRoute().Subrouter()})
}

// get is a helper method for a mux handler with a get method
func (r *router) get(path string, h http.HandlerFunc) *mux.Route {
	return r.HandleFunc(path, h).Methods("GET")
}

// pst is a helper method for a mux handler with a post method. Shortened to pst
// so that it lines up nicely with get when reading the routing code.
func (r *router) pst(path string, h http.HandlerFunc) *mux.Route {
	return r.HandleFunc(path, h).Methods("POST")
}
