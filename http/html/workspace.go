package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

func (app *Application) workspaceListRoute(organization string) string {
	return app.link("organizations", organization, "workspaces")
}
func (app *Application) workspaceListAnchor(organization string) anchor {
	return anchor{Name: "workspaces", Link: app.workspaceListRoute(organization)}
}
func (app *Application) workspaceListBreadcrumbs(organization string) []anchor {
	return append(app.siteBreadcrumbs(), app.workspaceListAnchor(organization))
}

func (app *Application) workspaceShowRoute(organization, workspace string) string {
	return app.link("organizations", organization, "workspaces", workspace)
}
func (app *Application) workspaceShowAnchor(organization, workspace string) anchor {
	return anchor{Name: workspace, Link: app.workspaceShowRoute(organization, workspace)}
}
func (app *Application) workspaceShowBreadcrumbs(organization, workspace string) []anchor {
	return append(app.workspaceListBreadcrumbs(organization), app.workspaceShowAnchor(organization, workspace))
}

func (app *Application) workspacesListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organization := vars["organization"]

	workspaces, err := app.WorkspaceService().List(r.Context(), otf.WorkspaceListOptions{OrganizationName: &organization})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(app.workspaceListBreadcrumbs(organization)...),
	}

	if err := app.render(r, "workspaces_list.tmpl", w, workspaces, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) workspacesShowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organization := vars["organization"]
	name := vars["workspace"]

	workspace, err := app.WorkspaceService().Get(r.Context(), otf.WorkspaceSpecifier{Name: &name})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(app.workspaceShowBreadcrumbs(organization, name)...),
	}

	if err := app.render(r, "workspaces_show.tmpl", w, workspace, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
