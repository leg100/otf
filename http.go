package otf

import "github.com/gorilla/mux"

// HTTPAPI is an HTTP API
type HTTPAPI interface {
	// AddHandlers adds http handlers for to the given mux.
	AddHandlers(*mux.Router)
}
