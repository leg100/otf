package html

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type WorkspaceController struct {
	otf.WorkspaceService

	// HTML template renderer
	renderer

	*templateDataFactory

	router *mux.Router
}

func (c *WorkspaceController) addRoutes(router *mux.Router) {
	router = router.PathPrefix("/organizations/{organization_name}/workspaces").Subrouter()

	router.HandleFunc("/", c.List).Methods("GET").Name("listWorkspace")
	router.HandleFunc("/new", c.New).Methods("GET").Name("newWorkspace")
	router.HandleFunc("/create", c.Create).Methods("POST").Name("createWorkspace")
	router.HandleFunc("/{workspace_name}", c.Get).Methods("GET").Name("getWorkspace")
	router.HandleFunc("/{workspace_name}/edit", c.Edit).Methods("GET").Name("editWorkspace")
	router.HandleFunc("/{workspace_name}/update", c.Update).Methods("POST").Name("updateWorkspace")
	router.HandleFunc("/{workspace_name}/delete", c.Delete).Methods("POST").Name("deleteWorkspace")
	router.HandleFunc("/{workspace_name}/editLock", c.EditLock).Methods("GET").Name("editWorkspace")
	router.HandleFunc("/{workspace_name}/updateLock", c.UpdateLock).Methods("POST").Name("lockWorkspace")
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

func (c *WorkspaceController) New(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceNewOptions
	if err := decode(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	tdata := c.newTemplateData(r, opts)

	if err := c.renderTemplate("workspaces_new.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *WorkspaceController) Create(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceCreateOptions
	if err := decode(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	workspace, err := c.WorkspaceService.Create(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, path.Join("..", workspace.Name), http.StatusFound)
}

func (c *WorkspaceController) Get(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpecifier

	if err := decode(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	workspace, err := c.WorkspaceService.Get(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, struct {
		Workspace *otf.Workspace
		Options   otf.WorkspaceSpecifier
	}{
		Workspace: workspace,
		Options:   opts,
	})

	if err := c.renderTemplate("workspaces_show.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *WorkspaceController) Edit(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpecifier

	if err := decode(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	workspace, err := c.WorkspaceService.Get(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, struct {
		Workspace *otf.Workspace
		Options   otf.WorkspaceSpecifier
	}{
		Workspace: workspace,
		Options:   opts,
	})

	if err := c.renderTemplate("workspaces_edit.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *WorkspaceController) Update(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceUpdateOptions
	if err := decode(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	workspace, err := c.WorkspaceService.Update(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "../edit", http.StatusFound)
}

func (c *WorkspaceController) Delete(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpecifier
	if err := decode(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	err := c.WorkspaceService.Delete(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "../../", http.StatusFound)
}
