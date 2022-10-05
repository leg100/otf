package html

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/r3labs/sse/v2"
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

func (app *Application) watchWorkspace(w http.ResponseWriter, r *http.Request) {
	streamID := r.URL.Query().Get("stream")
	if streamID == "" {
		writeError(w, "missing required query parameter: stream", http.StatusUnprocessableEntity)
		return
	}

	events, err := app.Watch(r.Context(), otf.WatchOptions{
		WorkspaceName:    otf.String(mux.Vars(r)["workspace_name"]),
		OrganizationName: otf.String(mux.Vars(r)["organization_name"]),
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				run, ok := event.Payload.(*otf.Run)
				if !ok {
					// skip non-run events
					continue
				}

				// Handle query parameters which filter run events: 'latest'
				// specifies that the client is only interested in event
				// concerning the latest run for the workspace; 'run-id' -
				// mutually exclusive with 'latest' - specifies that that the
				// client is only interested in events concerning the run with
				// that ID, and that is permitted to include a run which is
				// speculative (because web site shows speculative runs on the
				// run page but no where else); and if neither of those are
				// specified then events for all runs that are not speculative
				// are relayed (e.g. for web page that shows list of runs).
				if r.URL.Query().Get("latest") != "" {
					if !run.Latest() {
						// skip: run is not the latest run for a workspace
						continue
					}
				} else if runID := r.URL.Query().Get("run-id"); runID != "" {
					if runID != run.ID() {
						// skip: event is for a run which does not match the
						// filter
						continue
					}
				} else if run.Speculative() {
					// skip: event for speculative run
					continue
				}

				// render HTML snippets and send as payload in SSE event
				item := new(bytes.Buffer)
				if err := app.renderTemplate("run_item.tmpl", item, run); err != nil {
					app.Error(err, "rendering template for run item")
					continue
				}
				runStatus := new(bytes.Buffer)
				if err := app.renderTemplate("run_status.tmpl", runStatus, run); err != nil {
					app.Error(err, "rendering run status template")
					continue
				}
				planStatus := new(bytes.Buffer)
				if err := app.renderTemplate("phase_status.tmpl", planStatus, run.Plan()); err != nil {
					app.Error(err, "rendering plan status template")
					continue
				}
				applyStatus := new(bytes.Buffer)
				if err := app.renderTemplate("phase_status.tmpl", applyStatus, run.Apply()); err != nil {
					app.Error(err, "rendering apply status template")
					continue
				}
				js, err := json.Marshal(struct {
					ID string `json:"id"`
					// the run item box
					RunItem string `json:"html"`
					// the color-coded run status
					RunStatus string `json:"run-status"`
					// the color-coded plan status
					PlanStatus string `json:"plan-status"`
					// the color-coded apply status
					ApplyStatus string `json:"apply-status"`
				}{
					ID:          run.ID(),
					RunItem:     item.String(),
					RunStatus:   runStatus.String(),
					PlanStatus:  planStatus.String(),
					ApplyStatus: applyStatus.String(),
				})
				if err != nil {
					app.Error(err, "marshalling watched run", "run", run.ID())
					continue
				}
				app.Server.Publish(streamID, &sse.Event{
					Data:  js,
					Event: []byte(event.Type),
				})
			}
		}
	}()
	app.Server.ServeHTTP(w, r)
}
