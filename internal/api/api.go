// Package api provides commmon functionality for the OTF API
package api

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

type Handlers struct{}

func (h *Handlers) AddHandlers(r *mux.Router) {
	// basic no-op ping handler
	r.HandleFunc(path.Join(DefaultBasePath, "ping"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
