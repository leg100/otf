package html

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Router wraps mux's Router, adding various helper methods
type Router struct {
	*mux.Router
}

func NewRouter() *Router { return &Router{mux.NewRouter()} }

// Sub turns mux into chi (almost)
func (r *Router) Sub(group func(r *Router)) {
	group(&Router{r.NewRoute().Subrouter()})
}

// GET is a helper method for a mux handler with a GET method
func (r *Router) GET(path string, h http.HandlerFunc) *mux.Route {
	return r.HandleFunc(path, h).Methods("GET")
}

// PST is a helper method for a mux handler with a post method. Shortened to PST
// so that it lines up nicely with get when reading the routing code.
func (r *Router) PST(path string, h http.HandlerFunc) *mux.Route {
	return r.HandleFunc(path, h).Methods("POST")
}

// PTC is a helper method for a mux handler with a patch method. Shortened name
// so that it lines up nicely with get when reading the routing code.
func (r *Router) PTC(path string, h http.HandlerFunc) *mux.Route {
	return r.HandleFunc(path, h).Methods("PATCH")
}

// DEL is a helper method for a mux handler with a delete method. Shortened name
// so that it lines up nicely with get when reading the routing code.
func (r *Router) DEL(path string, h http.HandlerFunc) *mux.Route {
	return r.HandleFunc(path, h).Methods("DELETE")
}

// PUT is a helper method for a mux handler with a PUT method.
func (r *Router) PUT(path string, h http.HandlerFunc) *mux.Route {
	return r.HandleFunc(path, h).Methods("PUT")
}

func (r *Router) Headers(pairs ...string) *Route {
	return &Route{r.Router.Headers(pairs...)}
}

// Route wraps mux's Route, adding various helper methods
type Route struct {
	*mux.Route
}

// Sub turns mux into chi (almost)
func (r *Route) Sub(group func(r *Router)) {
	group(&Router{r.Subrouter()})
}

func (r *Route) PathPrefix(prefix string) *Route {
	return &Route{r.Route.PathPrefix(prefix)}
}
