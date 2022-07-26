package html

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

// workspaceRequest provides metadata about a request for a workspace
type workspaceRequest struct {
	r *http.Request
}

func (w workspaceRequest) OrganizationName() string {
	return param(w.r, "organization_name")
}

func (w workspaceRequest) WorkspaceName() string {
	return param(w.r, "workspace_name")
}

func (w workspaceRequest) Spec() otf.WorkspaceSpec {
	return otf.WorkspaceSpec{
		Name:             otf.String(w.WorkspaceName()),
		OrganizationName: otf.String(w.OrganizationName()),
	}
}

func (app *Application) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceListOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspaces, err := app.ListWorkspaces(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("workspace_list.tmpl", w, r, workspaceList{workspaces, opts})
}

func (app *Application) newWorkspace(w http.ResponseWriter, r *http.Request) {
	app.render("workspace_new.tmpl", w, r, organizationRequest{r})
}

func (app *Application) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspace, err := app.CreateWorkspace(r.Context(), opts)
	if err == otf.ErrResourcesAlreadyExists {
		flashError(w, "workspace already exists: "+opts.Name)
		http.Redirect(w, r, newWorkspacePath(organizationRequest{r}), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created workspace: "+workspace.Name())
	http.Redirect(w, r, getWorkspacePath(workspace), http.StatusFound)
}

func (app *Application) getWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	ws, err := app.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var latest *otf.Run
	if ws.LatestRunID() != nil {
		latest, err = app.GetRun(r.Context(), *ws.LatestRunID())
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	app.render("workspace_get.tmpl", w, r, struct {
		*otf.Workspace
		LatestRun      *otf.Run
		LatestStreamID string
	}{
		Workspace:      ws,
		LatestRun:      latest,
		LatestStreamID: "latest-" + otf.GenerateRandomString(5),
	})
}

func (app *Application) editWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	workspace, err := app.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("workspace_edit.tmpl", w, r, workspace)
}

func (app *Application) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var opts otf.WorkspaceUpdateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
	}
	if err := opts.Valid(); err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, editWorkspacePath(workspaceRequest{r}), http.StatusFound)
		return
	}
	workspace, err := app.UpdateWorkspace(r.Context(), spec, opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, editWorkspacePath(workspace), http.StatusFound)
}

func (app *Application) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err := app.DeleteWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted workspace: "+*spec.Name)
	http.Redirect(w, r, listWorkspacePath(organizationRequest{r}), http.StatusFound)
}

func (app *Application) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	ws, err := app.LockWorkspace(r.Context(), spec, otf.WorkspaceLockOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, getWorkspacePath(ws), http.StatusFound)
}

func (app *Application) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	ws, err := app.UnlockWorkspace(r.Context(), spec, otf.WorkspaceUnlockOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, getWorkspacePath(ws), http.StatusFound)
}
