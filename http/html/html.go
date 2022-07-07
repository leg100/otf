/*
Package html provides the oTF web app, serving up HTML formatted pages and associated assets (CSS, JS, etc).
*/
package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

const (
	defaultPageSize = 10
)

type pagination struct {
	*otf.Pagination
}

func param(r *http.Request, key string) string {
	vars := mux.Vars(r)
	if vars == nil {
		return ""
	}
	return vars[key]
}
