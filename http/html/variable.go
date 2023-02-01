package html

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
)

func (app *Application) newVariable(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("variable_new.tmpl", w, r, struct {
		Workspace  *otf.Workspace
		Variable   *otf.Variable
		EditMode   bool
		FormAction string
	}{
		Workspace:  ws,
		Variable:   &otf.Variable{},
		EditMode:   false,
		FormAction: paths.CreateVariable(workspaceID),
	})
}

func (app *Application) createVariable(w http.ResponseWriter, r *http.Request) {
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
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := app.CreateVariable(r.Context(), params.WorkspaceID, otf.CreateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   &params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	FlashSuccess(w, "added variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(params.WorkspaceID), http.StatusFound)
}

func (app *Application) listVariables(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variables, err := app.ListVariables(r.Context(), workspaceID)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := app.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("variable_list.tmpl", w, r, struct {
		Workspace *otf.Workspace
		Variables []*otf.Variable
	}{
		Workspace: ws,
		Variables: variables,
	})
}

func (app *Application) editVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := app.GetVariable(r.Context(), variableID)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := app.GetWorkspace(r.Context(), variable.WorkspaceID())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("variable_edit.tmpl", w, r, struct {
		Workspace  *otf.Workspace
		Variable   *otf.Variable
		EditMode   bool
		FormAction string
	}{
		Workspace:  ws,
		Variable:   variable,
		EditMode:   true,
		FormAction: paths.UpdateVariable(variable.ID()),
	})
}

func (app *Application) updateVariable(w http.ResponseWriter, r *http.Request) {
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
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// sensitive variable's value form field is deliberately empty, so avoid
	// updating value with empty string (this does mean users cannot set
	// a sensitive variable's value to an empty string but there unlikely to be
	// a valid reason to want to do that...)
	if params.Sensitive && params.Value != nil && *params.Value == "" {
		params.Value = nil
	}

	variable, err := app.UpdateVariable(r.Context(), params.VariableID, otf.UpdateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   &params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	FlashSuccess(w, "updated variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID()), http.StatusFound)
}

func (app *Application) deleteVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := app.DeleteVariable(r.Context(), variableID)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	FlashSuccess(w, "deleted variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID()), http.StatusFound)
}
