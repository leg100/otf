package html

import "github.com/gorilla/mux"

type router struct {
	*mux.Router
}

func (r *router) getRoute(name string, params ...string) string {
	url, err := r.Get(name).URLPath(params...)
	if err != nil {
		panic("no such web route exists: " + name)
	}

	return url.Path
}
