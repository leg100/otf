package state

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/resource"
)

type webHandlers struct {
	html.Renderer
	*Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/resources", h.getResources).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/outputs", h.getOutputs).Methods("GET")
}

type stateParams struct {
	resource.PageOptions

	WorkspaceID resource.ID `schema:"workspace_id,required"`
}

func (h *webHandlers) getResources(w http.ResponseWriter, r *http.Request) {
	var params stateParams
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// temp override number of items in a page to 10
	// params.PageSize = 10
	f, err := h.getCurrent(r.Context(), params.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// time.Sleep(time.Second)
	h.Render("state_resources.tmpl", w, html.Page[Resource]{
		Page:    resource.NewPage(f.Resources, params.PageOptions, nil),
		Request: r,
	})
}

func (h *webHandlers) getOutputs(w http.ResponseWriter, r *http.Request) {
	var params stateParams
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	f, err := h.getCurrent(r.Context(), params.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// convert outputs map to slice
	type output struct {
		Name  string
		Type  func() (string, error)
		Value string
	}
	outputs := make([]output, len(f.Outputs))
	var i int
	for name, out := range f.Outputs {
		outputs[i] = output{
			Name:  name,
			Type:  out.Type,
			Value: out.StringValue(),
		}
		i++
	}
	h.Render("state_outputs.tmpl", w, html.Page[output]{
		Page:    resource.NewPage(outputs, params.PageOptions, nil),
		Request: r,
	})
}

func (h *webHandlers) getCurrent(ctx context.Context, workspaceID resource.ID) (*File, error) {
	// ignore errors and instead render unpopulated template
	var f File
	sv, err := h.GetCurrent(ctx, workspaceID)
	if err != nil {
		return &f, nil
	}
	if err := json.Unmarshal(sv.State, &f); err != nil {
		return nil, err
	}
	return &f, nil
}
