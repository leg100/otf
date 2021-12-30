package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

var (
	organizationsAnchor = anchor{Name: "organizations", Link: "/organizations"}
	workspacesAnchor    = anchor{Name: "workspaces", Link: "/workspaces"}
)

func (app *Application) runsListHandler(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace_id"]

	runs, err := app.RunService().List(otf.RunListOptions{WorkspaceID: &workspaceID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(workspaceListAnchor, workspaceShowAnchor(workspaceID)),
	}

	if err := app.render(r, "runs_list.tmpl", w, runs, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) runsShowHandler(w http.ResponseWriter, r *http.Request) {
	runID := mux.Vars(r)["id"]

	runs, err := app.RunService().Get(runID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(workspaceListAnchor, workspaceShowAnchor(workspaceID)),
	}

	if err := app.render(r, "runs_list.tmpl", w, runs, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func runsShowBreadcrumbs(organization, workspace, runID string) []anchor {
}

func (app *Application) runListRoute() string { return app.link("/runs") }
func (app *Application) runListAnchor() anchor {
	return anchor{Name: "runs", Link: app.runListRoute()}
}
func (app *Application) runListBreadcrumbs() []anchor {
	return anchor{Name: "runs", Link: app.runListRoute()}
}

//func organizationsShowBreadcrumbs(organization string) []anchor { return
//append([]anchor{siteAnchor}, organizationsAnchor, anchor{Name: }
