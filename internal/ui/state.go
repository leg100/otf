package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/ui/helpers"
)

func addStateHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/workspaces/{workspace_id}/state", h.getState).Methods("GET")
}

func (h *Handlers) getState(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// ignore errors and instead render unpopulated template
	f := &state.File{}
	sv, err := h.State.GetCurrent(r.Context(), id)
	if err == nil {
		f, _ = sv.File()
	}

	helpers.Render(getState(f), w, r)
}
