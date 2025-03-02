package state

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
)

type webHandlers struct {
	*Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/state", h.getState).Methods("GET")
}

func (h *webHandlers) getState(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// ignore errors and instead render unpopulated template
	f := &File{}
	sv, err := h.GetCurrent(r.Context(), id)
	if err == nil {
		f, _ = sv.File()
	}

	html.Render(get(f), w, r)
}
