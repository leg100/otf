package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

var workspaceListAnchor = anchor{
	Name: "workspaces",
	Link: "/workspaces",
}

func (app *Application) workspacesListHandler(w http.ResponseWriter, r *http.Request) {
	workspaces, err := app.WorkspaceService().List(r.Context(), otf.WorkspaceListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(workspaceListAnchor),
	}

	if err := app.render(r, "workspaces_list.tmpl", w, workspaces, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) workspacesShowHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	workspace, err := app.WorkspaceService().Get(r.Context(), otf.WorkspaceSpecifier{ID: &id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(workspaceListAnchor, workspaceShowAnchor(id)),
	}

	if err := app.render(r, "workspaces_show.tmpl", w, workspace, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func workspaceShowAnchor(id string) anchor {
	return anchor{
		Name: id,
		Link: "/workspaces/" + id,
	}
}
