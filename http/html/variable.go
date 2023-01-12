package html

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
)

type variableForm struct {
	Action string
	*otf.Variable
}

func (app *Application) newVariable(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), otf.WorkspaceSpec{ID: otf.String(workspaceID)})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("variable_new.tmpl", w, r, struct {
		Workspace  *otf.Workspace
		Variable   *otf.Variable
		FormAction string
	}{
		Workspace:  ws,
		Variable:   &otf.Variable{},
		FormAction: paths.CreateVariable(workspaceID),
	})
}

func (app *Application) createVariable(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		otf.CreateVariableOptions
		WorkspaceID string `schema:"workspace_id,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := app.CreateVariable(r.Context(), params.WorkspaceID, params.CreateVariableOptions)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "added variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(params.WorkspaceID), http.StatusFound)
}

func (app *Application) listVariables(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variables, err := app.ListVariables(r.Context(), workspaceID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := app.GetWorkspace(r.Context(), otf.WorkspaceSpec{ID: otf.String(workspaceID)})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("variable_list.tmpl", w, r, struct {
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
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := app.GetVariable(r.Context(), variableID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := app.GetWorkspace(r.Context(), otf.WorkspaceSpec{ID: otf.String(variable.WorkspaceID())})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("variable_edit.tmpl", w, r, struct {
		Workspace  *otf.Workspace
		Variable   *otf.Variable
		FormAction string
	}{
		Workspace:  ws,
		Variable:   variable,
		FormAction: paths.UpdateVariable(variable.ID()),
	})
}

func (app *Application) updateVariable(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		otf.UpdateVariableOptions
		VariableID string `schema:"variable_id,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := app.UpdateVariable(r.Context(), params.VariableID, params.UpdateVariableOptions)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "updated variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID()), http.StatusFound)
}

func (app *Application) deleteVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := app.DeleteVariable(r.Context(), variableID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "deleted variable: "+variable.Key())
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID()), http.StatusFound)
}
