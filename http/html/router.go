package html

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type router struct {
	*mux.Router
}

// relative gets a relative URL path with the given route name. Route variables
// are retrieved from the current request and further variables can be provided
// as parameters.
func (r *router) relative(req *http.Request, name string, params ...string) string {
	reqVars := flattenMap(mux.Vars(req))

	return r.route(name, append(reqVars, params...)...)
}

func (r *router) route(name string, params ...string) string {
	route := r.Get(name)

	if route == nil {
		panic(fmt.Errorf("no such web route exists: %s", name))
	}

	url, err := route.URLPath(params...)
	if err != nil {
		panic(err)
	}

	return url.Path
}
