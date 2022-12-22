package html

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
)

func (app *Application) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		TeamName    string `schema:"team_name,required"`
		Role        string `schema:"role,required"`
	}
	params := parameters{}
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	role, err := otf.WorkspaceRoleFromString(params.Role)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	spec := otf.WorkspaceSpec{ID: otf.String(params.WorkspaceID)}
	ws, err := app.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.SetWorkspacePermission(r.Context(), spec, params.TeamName, role)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "updated workspace permissions")
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *Application) unsetWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		TeamName    string `schema:"team_name,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	spec := otf.WorkspaceSpec{ID: otf.String(params.WorkspaceID)}
	ws, err := app.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.UnsetWorkspacePermission(r.Context(), spec, params.TeamName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted workspace permission")
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}
