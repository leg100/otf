/*
Package html provides the otf web app, serving up HTML formatted pages and associated assets (CSS, JS, etc).
*/
package html

import (
	"net/http"

	"github.com/gorilla/mux"
)

// ListOptions is used to specify pagination options when making HTTP requests.
// Pagination allows breaking up large result sets into chunks, or "pages".
type ListOptions struct {
	// The page number to request. The results vary based on the PageSize.
	PageNumber int `schema:"page[number],omitempty"`
	// The number of elements returned in a single page.
	PageSize int `schema:"page[size],omitempty"`
}

func param(r *http.Request, key string) string {
	vars := mux.Vars(r)
	if vars == nil {
		return ""
	}
	return vars[key]
}
