// Package api provides support for the OTF API
package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
)

const (
	BasePath     = "/otfapi"
	PingEndpoint = "ping"
)

type Handlers struct {
	Handlers []internal.Handlers
}

func (h *Handlers) AddHandlers(r *mux.Router) {
	r = r.PathPrefix(BasePath).Subrouter()

	// basic no-op ping handler for API
	r.HandleFunc(PingEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	for _, h := range h.Handlers {
		h.AddHandlers(r)
	}
}
