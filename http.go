package otf

import "github.com/gorilla/mux"

// Handlers is an http application with handlers
type Handlers interface {
	// AddHandlers adds http handlers to the router.
	AddHandlers(*mux.Router)
}
