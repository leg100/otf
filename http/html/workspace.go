package html

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
	"github.com/r3labs/sse/v2"
)

func (app *Application) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceListOptions
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	org, err := app.GetOrganization(r.Context(), *opts.Organization)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	workspaces, err := app.ListWorkspaces(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("workspace_list.tmpl", w, r, struct {
		*otf.WorkspaceList
		*otf.Organization
	}{
		WorkspaceList: workspaces,
		Organization:  org,
	})
}

func (app *Application) newWorkspace(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := app.GetOrganization(r.Context(), organization)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("workspace_new.tmpl", w, r, org)
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

	org, err := app.GetOrganization(r.Context(), opts.Organization)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	workspace, err := app.CreateWorkspace(r.Context(), opts)
	if err == otf.ErrResourceAlreadyExists {
		flashError(w, "workspace already exists: "+opts.Name)
		http.Redirect(w, r, paths.NewWorkspace(org.Name()), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created workspace: "+workspace.Name())
	http.Redirect(w, r, paths.Workspace(workspace.ID()), http.StatusFound)
}

func (app *Application) getWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.All(&spec, r); err != nil {
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
	// Get existing perms as well as all teams in org
	perms, err := app.ListWorkspacePermissions(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	teams, err := app.ListTeams(r.Context(), workspace.Organization())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Filter teams, removing those already assigned perms
	for _, p := range perms {
		for it, t := range teams {
			if t.ID() == p.Team.ID() {
				teams = append(teams[:it], teams[it+1:]...)
				break
			}
		}
	}

	app.render("workspace_edit.tmpl", w, r, struct {
		*otf.Workspace
		Permissions []*otf.WorkspacePermission
		Teams       []*otf.Team
		Roles       []otf.Role
	}{
		Workspace:   workspace,
		Permissions: perms,
		Teams:       teams,
		Roles: []otf.Role{
			otf.WorkspaceReadRole,
			otf.WorkspacePlanRole,
			otf.WorkspaceWriteRole,
			otf.WorkspaceAdminRole,
		},
	})
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
	ws, err := app.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := opts.Valid(); err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, paths.EditWorkspace(ws.ID()), http.StatusFound)
		return
	}

	// TODO: add support for updating vcs repo, e.g. branch, etc.
	ws, err = app.UpdateWorkspace(r.Context(), spec, opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, paths.EditWorkspace(ws.ID()), http.StatusFound)
}

func (app *Application) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.DeleteWorkspace(r.Context(), otf.WorkspaceSpec{ID: otf.String(id)})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted workspace: "+ws.Name())
	http.Redirect(w, r, paths.Workspaces(ws.Organization()), http.StatusFound)
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
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
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
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *Application) watchWorkspace(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		StreamID    string `schema:"stream,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), otf.WorkspaceSpec{ID: &params.WorkspaceID})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	events, err := app.Watch(r.Context(), otf.WatchOptions{
		WorkspaceID: otf.String(ws.ID()),
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
				itemHTML := new(bytes.Buffer)
				if err := app.renderTemplate("run_item.tmpl", itemHTML, run); err != nil {
					app.Error(err, "rendering template for run item")
					continue
				}
				runStatusHTML := new(bytes.Buffer)
				if err := app.renderTemplate("run_status.tmpl", runStatusHTML, run); err != nil {
					app.Error(err, "rendering run status template")
					continue
				}
				planStatusHTML := new(bytes.Buffer)
				if err := app.renderTemplate("phase_status.tmpl", planStatusHTML, run.Plan()); err != nil {
					app.Error(err, "rendering plan status template")
					continue
				}
				applyStatusHTML := new(bytes.Buffer)
				if err := app.renderTemplate("phase_status.tmpl", applyStatusHTML, run.Apply()); err != nil {
					app.Error(err, "rendering apply status template")
					continue
				}
				js, err := json.Marshal(struct {
					ID              string        `json:"id"`
					RunStatus       otf.RunStatus `json:"run-status"`
					RunItemHTML     string        `json:"run-item-html"`
					RunStatusHTML   string        `json:"run-status-html"`
					PlanStatusHTML  string        `json:"plan-status-html"`
					ApplyStatusHTML string        `json:"apply-status-html"`
				}{
					ID:              run.ID(),
					RunStatus:       run.Status(),
					RunItemHTML:     itemHTML.String(),
					RunStatusHTML:   runStatusHTML.String(),
					PlanStatusHTML:  planStatusHTML.String(),
					ApplyStatusHTML: applyStatusHTML.String(),
				})
				if err != nil {
					app.Error(err, "marshalling watched run", "run", run.ID())
					continue
				}
				app.Server.Publish(params.StreamID, &sse.Event{
					Data:  js,
					Event: []byte(event.Type),
				})
			}
		}
	}()
	app.Server.ServeHTTP(w, r)
}

func (app *Application) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), otf.WorkspaceSpec{ID: &workspaceID})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	providers, err := app.ListVCSProviders(r.Context(), ws.Organization())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("workspace_vcs_provider_list.tmpl", w, r, struct {
		Items []*otf.VCSProvider
		*otf.Workspace
	}{
		Items:     providers,
		Workspace: ws,
	})
}

func (app *Application) listWorkspaceVCSRepos(w http.ResponseWriter, r *http.Request) {
	type options struct {
		WorkspaceID     string `schema:"workspace_id,required"`
		VCSProviderID   string `schema:"vcs_provider_id,required"`
		otf.ListOptions        // Pagination
		// TODO: filters, public/private, etc
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), otf.WorkspaceSpec{ID: &opts.WorkspaceID})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	provider, err := app.GetVCSProvider(r.Context(), opts.VCSProviderID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repos, err := app.ListRepositories(r.Context(), opts.VCSProviderID, cloud.ListRepositoriesOptions{
		PageSize: 100,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("workspace_vcs_repo_list.tmpl", w, r, struct {
		Items []cloud.Repo
		*otf.Workspace
		Provider *otf.VCSProvider
	}{
		Items:     repos,
		Workspace: ws,
		Provider:  provider,
	})
}

func (app *Application) connectWorkspace(w http.ResponseWriter, r *http.Request) {
	type options struct {
		otf.WorkspaceSpec
		otf.ConnectWorkspaceOptions
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.GetVCSProvider(r.Context(), opts.ProviderID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	opts.Cloud = provider.CloudConfig().Name

	ws, err := app.ConnectWorkspace(r.Context(), opts.WorkspaceSpec, opts.ConnectWorkspaceOptions)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *Application) disconnectWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.All(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.DisconnectWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *Application) startRun(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		Strategy    string `schema:"strategy,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var speculative bool
	switch params.Strategy {
	case "plan-only":
		speculative = true
	case "plan-and-apply":
		speculative = false
	default:
		writeError(w, "invalid strategy", http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), otf.WorkspaceSpec{ID: &params.WorkspaceID})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	run, err := app.StartRun(r.Context(), otf.WorkspaceSpec{ID: &params.WorkspaceID}, otf.ConfigurationVersionCreateOptions{
		Speculative: otf.Bool(speculative),
	})
	if err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID()), http.StatusFound)
}
