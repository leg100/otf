package ui

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/ui/paths"
)

// Handlers is the collection of UI handlers
type Handlers struct {
	Handlers []internal.Handlers
}

func (h *Handlers) AddHandlers(r *mux.Router) {
	r = r.PathPrefix(paths.UIPrefix).Subrouter()
	for _, h := range h.Handlers {
		h.AddHandlers(r)
	}
}
