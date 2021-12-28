package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
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
