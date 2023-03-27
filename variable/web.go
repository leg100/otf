package variable

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/workspace"
)

type web struct {
	otf.Renderer
	workspace.Service

	svc Service
}

func (h *web) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/variables", h.list)
	r.HandleFunc("/workspaces/{workspace_id}/variables/new", h.new)
	r.HandleFunc("/workspaces/{workspace_id}/variables/create", h.create)
	r.HandleFunc("/variables/{variable_id}/edit", h.edit)
	r.HandleFunc("/variables/{variable_id}/update", h.update)
	r.HandleFunc("/variables/{variable_id}/delete", h.delete)
}

func (h *web) new(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("variable_new.tmpl", w, r, struct {
		Workspace  *workspace.Workspace
		Variable   *Variable
		EditMode   bool
		FormAction string
	}{
		Workspace:  ws,
		Variable:   &Variable{},
		EditMode:   false,
		FormAction: paths.CreateVariable(workspaceID),
	})
}

func (h *web) create(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Key         *string `schema:"key,required"`
		Value       *string
		Description *string
		Category    *VariableCategory `schema:"category,required"`
		Sensitive   bool
		HCL         bool
		WorkspaceID string `schema:"workspace_id,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.CreateVariable(r.Context(), params.WorkspaceID, CreateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   &params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "added variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(params.WorkspaceID), http.StatusFound)
}

func (h *web) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variables, err := h.svc.ListVariables(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("variable_list.tmpl", w, r, struct {
		Workspace *workspace.Workspace
		Variables []*Variable
	}{
		Workspace: ws,
		Variables: variables,
	})
}

func (h *web) edit(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.GetVariable(r.Context(), variableID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.GetWorkspace(r.Context(), variable.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("variable_edit.tmpl", w, r, struct {
		Workspace  *workspace.Workspace
		Variable   *Variable
		EditMode   bool
		FormAction string
	}{
		Workspace:  ws,
		Variable:   variable,
		EditMode:   true,
		FormAction: paths.UpdateVariable(variable.ID),
	})
}

func (h *web) update(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Key         *string `schema:"key,required"`
		Value       *string
		Description *string
		Category    *VariableCategory `schema:"category,required"`
		Sensitive   bool
		HCL         bool
		VariableID  string `schema:"variable_id,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// sensitive variable's value form field is deliberately empty, so avoid
	// updating value with empty string (this does mean users cannot set
	// a sensitive variable's value to an empty string but there unlikely to be
	// a valid reason to want to do that...)
	if params.Sensitive && params.Value != nil && *params.Value == "" {
		params.Value = nil
	}

	variable, err := h.svc.UpdateVariable(r.Context(), params.VariableID, UpdateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   &params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID), http.StatusFound)
}

func (h *web) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.DeleteVariable(r.Context(), variableID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID), http.StatusFound)
}
