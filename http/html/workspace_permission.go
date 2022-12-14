package html

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (app *Application) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	type workspacePermission struct {
		Name             string `schema:"workspace_name,required"`
		OrganizationName string `schema:"organization_name,required"`
		TeamName         string `schema:"team_name,required"`
		Role             string `schema:"role,required"`
	}
	perm := workspacePermission{}
	if err := decode.All(&perm, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	role, err := otf.WorkspaceRoleFromString(perm.Role)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	spec := otf.WorkspaceSpec{Name: otf.String(perm.Name), OrganizationName: otf.String(perm.OrganizationName)}
	ws, err := app.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.SetWorkspacePermission(r.Context(), spec, perm.TeamName, role)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "updated workspace permissions")
	http.Redirect(w, r, getWorkspacePath(ws), http.StatusFound)
}

func (app *Application) unsetWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	type workspacePermission struct {
		Name             string `schema:"workspace_name,required"`
		OrganizationName string `schema:"organization_name,required"`
		TeamName         string `schema:"team_name,required"`
	}
	perm := workspacePermission{}
	if err := decode.All(&perm, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	spec := otf.WorkspaceSpec{Name: otf.String(perm.Name), OrganizationName: otf.String(perm.OrganizationName)}
	ws, err := app.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.UnsetWorkspacePermission(r.Context(), spec, perm.TeamName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted workspace permission")
	http.Redirect(w, r, getWorkspacePath(ws), http.StatusFound)
}
