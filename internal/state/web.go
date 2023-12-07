package state

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
)

type webHandlers struct {
	html.Renderer
	*Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/state", h.getState).Methods("GET")
}

func (h *webHandlers) getState(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// ignore errors and instead render unpopulated template
	f := &File{}
	sv, err := h.GetCurrentStateVersion(r.Context(), id)
	if err == nil {
		f, _ = sv.File()
	}

	if err := h.RenderTemplate("state_get.tmpl", w, f); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
