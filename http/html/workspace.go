package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

type WorkspaceController struct {
	otf.WorkspaceService

	// HTML template renderer
	renderer

	*router

	// for setting flash messages
	sessions *sessions

	*templateDataFactory
}

func (c *WorkspaceController) addRoutes(router *mux.Router) {
	router.HandleFunc("/", c.List).Methods("GET").Name("listWorkspace")
	router.HandleFunc("/new", c.New).Methods("GET").Name("newWorkspace")
	router.HandleFunc("/create", c.Create).Methods("POST").Name("createWorkspace")
	router.HandleFunc("/{workspace_name}", c.Get).Methods("GET").Name("getWorkspace")
	router.HandleFunc("/{workspace_name}/edit", c.Edit).Methods("GET").Name("editWorkspace")
	router.HandleFunc("/{workspace_name}/update", c.Update).Methods("POST").Name("updateWorkspace")
	router.HandleFunc("/{workspace_name}/delete", c.Delete).Methods("POST").Name("deleteWorkspace")
	router.HandleFunc("/{workspace_name}/lock", c.Lock).Methods("POST").Name("lockWorkspace")
	router.HandleFunc("/{workspace_name}/unlock", c.Unlock).Methods("POST").Name("unlockWorkspace")
}

func (c *WorkspaceController) List(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceListOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspaces, err := c.WorkspaceService.List(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tdata := c.newTemplateData(r, struct {
		List    *otf.WorkspaceList
		Options otf.WorkspaceListOptions
	}{workspaces, opts})
	if err := c.renderTemplate("workspace_list.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *WorkspaceController) New(w http.ResponseWriter, r *http.Request) {
	tdata := c.newTemplateData(r, mux.Vars(r)["organization_name"])
	if err := c.renderTemplate("workspace_new.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}
func (c *WorkspaceController) Create(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspace, err := c.WorkspaceService.Create(r.Context(), opts)
	if err == otf.ErrResourcesAlreadyExists {
		c.sessions.FlashError(r, "workspace already exists: ", opts.Name)
		http.Redirect(w, r, c.relative(r, "newWorkspace"), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.sessions.FlashSuccess(r, "created workspace: ", workspace.Name())
	http.Redirect(w, r, c.relative(r, "getWorkspace", "workspace_name", opts.Name), http.StatusFound)
}

func (c *WorkspaceController) Get(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpec
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspace, err := c.WorkspaceService.Get(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tdata := c.newTemplateData(r, workspace)
	if err := c.renderTemplate("workspace_get.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *WorkspaceController) Edit(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpec
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspace, err := c.WorkspaceService.Get(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, struct {
		Workspace *otf.Workspace
		Options   otf.WorkspaceSpec
	}{
		Workspace: workspace,
		Options:   opts,
	})

	if err := c.renderTemplate("workspace_edit.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *WorkspaceController) Update(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var opts otf.WorkspaceUpdateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
	}
	workspace, err := c.WorkspaceService.Update(r.Context(), spec, opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.sessions.FlashSuccess(r, "updated workspace")

	// Explicitly specify route variables because user may have updated them.
	http.Redirect(w, r, c.relative(r, "editWorkspace", "workspace_name", workspace.Name()), http.StatusFound)
}

func (c *WorkspaceController) Delete(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpec
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err := c.WorkspaceService.Delete(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.sessions.FlashSuccess(r, "deleted workspace: ", *opts.Name)
	http.Redirect(w, r, c.relative(r, "listWorkspace"), http.StatusFound)
}

func (c *WorkspaceController) Lock(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, err := c.WorkspaceService.Lock(r.Context(), spec, otf.WorkspaceLockOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, c.relative(r, "getWorkspace"), http.StatusFound)
}

func (c *WorkspaceController) Unlock(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, err := c.WorkspaceService.Unlock(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, c.relative(r, "getWorkspace"), http.StatusFound)
}
