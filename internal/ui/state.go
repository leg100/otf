package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/state"
)

type webHandlers struct {
	*state.Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/state", h.getState).Methods("GET")
}

func (h *webHandlers) getState(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// ignore errors and instead render unpopulated template
	f := &state.File{}
	sv, err := h.GetCurrent(r.Context(), id)
	if err == nil {
		f, _ = sv.File()
	}

	html.Render(getState(f), w, r)
}
