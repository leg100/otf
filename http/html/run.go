package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

func (app *Application) runListRoute(organization, workspace string) string {
	return app.link("organizations", organization, "workspaces", workspace, "runs")
}
func (app *Application) runListAnchor(organization, workspace string) anchor {
	return anchor{Name: "runs", Link: app.runListRoute(organization, workspace)}
}
func (app *Application) runListBreadcrumbs(organization, workspace string) []anchor {
	return append(app.workspaceShowBreadcrumbs(organization, workspace), app.runListAnchor(organization, workspace))
}

func (app *Application) runShowRoute(organization, workspace, runID string) string {
	return app.link("organizations", organization, "workspaces", workspace, "runs", runID)
}
func (app *Application) runShowAnchor(organization, workspace, runID string) anchor {
	return anchor{Name: "runs", Link: app.runShowRoute(organization, workspace, runID)}
}
func (app *Application) runShowBreadcrumbs(organization, workspace, runID string) []anchor {
	return append(app.runListBreadcrumbs(organization, workspace), app.runShowAnchor(organization, workspace, runID))
}

//func organizationsShowBreadcrumbs(organization string) []anchor { return
//append([]anchor{siteAnchor}, organizationsAnchor, anchor{Name: }
func (app *Application) runsListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organization := vars["organization"]
	workspace := vars["workspace"]

	runs, err := app.RunService().List(otf.RunListOptions{OrganizationName: &organization, WorkspaceName: &workspace})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(app.runListBreadcrumbs(organization, workspace)...),
	}

	if err := app.render(r, "runs_list.tmpl", w, runs, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) runsShowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organization := vars["organization"]
	workspace := vars["workspace"]
	runID := vars["id"]

	run, err := app.RunService().Get(runID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(app.runShowBreadcrumbs(organization, workspace, runID)...),
	}

	if err := app.render(r, "runs_show.tmpl", w, run, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
