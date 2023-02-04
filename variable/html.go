package variable

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

type htmlApp struct {
	otf.Renderer
	otf.WorkspaceService

	app service
}

func (a *htmlApp) AddHTMLHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/variables", a.list)
	r.HandleFunc("/workspaces/{workspace_id}/variables/new", a.new)
	r.HandleFunc("/workspaces/{workspace_id}/variables/create", a.create)
	r.HandleFunc("/variables/{variable_id}/edit", a.edit)
	r.HandleFunc("/variables/{variable_id}/update", a.update)
	r.HandleFunc("/variables/{variable_id}/delete", a.delete)
}

func (a *htmlApp) new(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := a.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("variable_new.tmpl", w, r, struct {
		Workspace  *otf.Workspace
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

func (a *htmlApp) create(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Key         *string `schema:"key,required"`
		Value       *string
		Description *string
		Category    *otf.VariableCategory `schema:"category,required"`
		Sensitive   bool
		HCL         bool
		WorkspaceID string `schema:"workspace_id,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := a.app.create(r.Context(), params.WorkspaceID, otf.CreateVariableOptions{
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

	html.FlashSuccess(w, "added variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(params.WorkspaceID), http.StatusFound)
}

func (a *htmlApp) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variables, err := a.app.list(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := a.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("variable_list.tmpl", w, r, struct {
		Workspace *otf.Workspace
		Variables []*Variable
	}{
		Workspace: ws,
		Variables: variables,
	})
}

func (a *htmlApp) edit(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := a.app.get(r.Context(), variableID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := a.GetWorkspace(r.Context(), variable.WorkspaceID())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("variable_edit.tmpl", w, r, struct {
		Workspace  *otf.Workspace
		Variable   *Variable
		EditMode   bool
		FormAction string
	}{
		Workspace:  ws,
		Variable:   variable,
		EditMode:   true,
		FormAction: paths.UpdateVariable(variable.ID()),
	})
}

func (a *htmlApp) update(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Key         *string `schema:"key,required"`
		Value       *string
		Description *string
		Category    *otf.VariableCategory `schema:"category,required"`
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

	variable, err := a.app.update(r.Context(), params.VariableID, otf.UpdateVariableOptions{
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

	html.FlashSuccess(w, "updated variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID()), http.StatusFound)
}

func (a *htmlApp) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := a.app.delete(r.Context(), variableID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID()), http.StatusFound)
}
