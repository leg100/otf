package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type WorkspaceController struct {
	otf.WorkspaceService

	// HTML template renderer
	renderer

	*templateDataFactory
}

func (c *WorkspaceController) addRoutes(router *mux.Router) {
	router = router.PathPrefix("/organizations/{organization_name}/workspaces").Subrouter()

	router.Handle("/", c.List).Methods("GET").Name("listWorkspace")
	router.Handle("/new", c.New).Methods("GET").Name("newWorkspace")
	router.Handle("/create", c.Create).Methods("POST").Name("createWorkspace")
	router.Handle("/{workspace_name}", c.Get).Methods("GET").Name("showWorkspace")
	router.Handle("/{workspace_name}/edit", c.Edit).Methods("GET").Name("editWorkspace")
	router.Handle("/{workspace_name}/update", c.Update).Methods("POST").Name("updateWorkspace")
	router.Handle("/{workspace_name}/delete", c.Delete).Methods("POST").Name("deleteWorkspace")
}

func (c *WorkspaceController) List(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceListOptions

	// populate options struct from query and route paramters
	if err := decode(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	workspaces, err := c.WorkspaceService.List(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, struct {
		List    *otf.WorkspaceList
		Options otf.WorkspaceListOptions
	}{
		List:    workspaces,
		Options: opts,
	})

	if err := c.renderTemplate("workspaces_list.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) workspacesShowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organization := vars["organization"]
	name := vars["name"]

	workspace, err := app.WorkspaceService().Get(r.Context(), otf.WorkspaceSpecifier{OrganizationName: &organization, Name: &name})
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

func (app *Application) workspaceListRoute(organization string) string {
	return app.link("organizations", organization, "workspaces")
}
func (app *Application) workspaceListAnchor(organization string) anchor {
	return anchor{Name: "workspaces", Link: app.workspaceListRoute(organization)}
}
func (app *Application) workspaceListBreadcrumbs(organization string) []anchor {
	return append(app.organizationShowBreadcrumbs(organization), app.workspaceListAnchor(organization))
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
