package html

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/r3labs/sse/v2"
)

func (app *Application) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceListOptions
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	org, err := app.GetOrganization(r.Context(), *opts.OrganizationName)
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

	org, err := app.GetOrganization(r.Context(), opts.OrganizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	workspace, err := app.CreateWorkspace(r.Context(), opts)
	if err == otf.ErrResourceAlreadyExists {
		flashError(w, "workspace already exists: "+opts.Name)
		http.Redirect(w, r, newWorkspacePath(org.Name()), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created workspace: "+workspace.Name())
	http.Redirect(w, r, workspacePath(workspace.ID()), http.StatusFound)
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

	// Get existing perms as well as all teams in org
	perms, err := app.ListWorkspacePermissions(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	teams, err := app.ListTeams(r.Context(), ws.OrganizationName())
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

	app.render("workspace_get.tmpl", w, r, struct {
		*otf.Workspace
		LatestRun      *otf.Run
		LatestStreamID string
		Permissions    []*otf.WorkspacePermission
		Teams          []*otf.Team
		Roles          []otf.Role
	}{
		Workspace:      ws,
		LatestRun:      latest,
		LatestStreamID: "latest-" + otf.GenerateRandomString(5),
		Permissions:    perms,
		Teams:          teams,
		Roles: []otf.Role{
			otf.WorkspaceReadRole,
			otf.WorkspacePlanRole,
			otf.WorkspaceWriteRole,
			otf.WorkspaceAdminRole,
		},
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
	ws, err := app.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := opts.Valid(); err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, editWorkspacePath(ws.ID()), http.StatusFound)
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
	http.Redirect(w, r, editWorkspacePath(ws.ID()), http.StatusFound)
}

func (app *Application) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := app.GetOrganization(r.Context(), *spec.OrganizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.DeleteWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted workspace: "+*spec.Name)
	http.Redirect(w, r, workspacesPath(org.Name()), http.StatusFound)
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
	http.Redirect(w, r, workspacePath(ws.ID()), http.StatusFound)
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
	http.Redirect(w, r, workspacePath(ws.ID()), http.StatusFound)
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
				app.Server.Publish(streamID, &sse.Event{
					Data:  js,
					Event: []byte(event.Type),
				})
			}
		}
	}()
	app.Server.ServeHTTP(w, r)
}

func (app *Application) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
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
	providers, err := app.ListVCSProviders(r.Context(), mux.Vars(r)["organization_name"])
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
		OrganizationName string `schema:"organization_name,required"`
		WorkspaceName    string `schema:"workspace_name,required"`
		VCSProviderID    string `schema:"vcs_provider_id,required"`
		// Pagination
		otf.ListOptions
		// TODO: filters, public/private, etc
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), otf.WorkspaceSpec{
		Name:             &opts.WorkspaceName,
		OrganizationName: &opts.OrganizationName,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	provider, err := app.GetVCSProvider(r.Context(), opts.VCSProviderID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repos, err := app.ListRepositories(r.Context(), opts.VCSProviderID, opts.ListOptions)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("workspace_vcs_repo_list.tmpl", w, r, struct {
		*otf.RepoList
		*otf.Workspace
		Provider *otf.VCSProvider
	}{
		RepoList:  repos,
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

	// extract externally-accessible host from request
	opts.OTFHost = otfhttp.ExternalHost(r)

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
	http.Redirect(w, r, workspacePath(ws.ID()), http.StatusFound)
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
	http.Redirect(w, r, workspacePath(ws.ID()), http.StatusFound)
}

func (app *Application) startRun(w http.ResponseWriter, r *http.Request) {
	// TODO: set cv opts directly, populating speculative parameter rather a new
	// strategy parameter.
	type options struct {
		otf.WorkspaceSpec
		Strategy string `schema:"strategy,required"`
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var speculative bool
	switch opts.Strategy {
	case "plan-only":
		speculative = true
	case "plan-and-apply":
		speculative = false
	default:
		writeError(w, "invalid strategy", http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), opts.WorkspaceSpec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	run, err := app.StartRun(r.Context(), opts.WorkspaceSpec, otf.ConfigurationVersionCreateOptions{
		Speculative: otf.Bool(speculative),
	})
	if err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, workspacePath(ws.ID()), http.StatusFound)
		return
	}

	http.Redirect(w, r, runPath(run.ID()), http.StatusFound)
}
