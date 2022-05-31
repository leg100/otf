package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (c *Application) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceListOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspaces, err := c.WorkspaceService().List(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.render("workspace_list.tmpl", w, r, struct {
		List    *otf.WorkspaceList
		Options otf.WorkspaceListOptions
	}{workspaces, opts})
}

func (c *Application) newWorkspace(w http.ResponseWriter, r *http.Request) {
	c.render("workspace_new.tmpl", w, r, mux.Vars(r)["organization_name"])
}

func (c *Application) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspace, err := c.WorkspaceService().Create(r.Context(), opts)
	if err == otf.ErrResourcesAlreadyExists {
		flashError(w, "workspace already exists: "+opts.Name)
		http.Redirect(w, r, c.relative(r, "newWorkspace"), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created workspace: "+workspace.Name())
	http.Redirect(w, r, c.relative(r, "getWorkspace", "workspace_name", opts.Name), http.StatusFound)
}

func (c *Application) getWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpec
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspace, err := c.WorkspaceService().Get(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.render("workspace_get.tmpl", w, r, workspace)
}

func (c *Application) editWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpec
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspace, err := c.WorkspaceService().Get(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.render("workspace_edit.tmpl", w, r, struct {
		Workspace *otf.Workspace
		Options   otf.WorkspaceSpec
	}{
		Workspace: workspace,
		Options:   opts,
	})
}

func (c *Application) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var opts otf.WorkspaceUpdateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
	}
	workspace, err := c.WorkspaceService().Update(r.Context(), spec, opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "updated workspace")
	// Explicitly specify route variables because user may have updated them.
	http.Redirect(w, r, c.relative(r, "editWorkspace", "workspace_name", workspace.Name()), http.StatusFound)
}

func (c *Application) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceSpec
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err := c.WorkspaceService().Delete(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted workspace: "+*opts.Name)
	http.Redirect(w, r, c.relative(r, "listWorkspace"), http.StatusFound)
}

func (c *Application) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	user, err := getCtxUser(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, err = c.WorkspaceService().Lock(r.Context(), spec, otf.WorkspaceLockOptions{
		Requestor: user,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, c.relative(r, "getWorkspace"), http.StatusFound)
}

func (c *Application) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	user, err := getCtxUser(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, err = c.WorkspaceService().Unlock(r.Context(), spec, otf.WorkspaceUnlockOptions{
		Requestor: user,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, c.relative(r, "getWorkspace"), http.StatusFound)
}
