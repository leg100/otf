package html

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/r3labs/sse/v2"
	"golang.org/x/oauth2"
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

// vcsProviderRequest provides metadata about a request for a vcs provider
type vcsProviderRequest struct {
	workspaceRequest
}

func (r vcsProviderRequest) VCSProviderID() string {
	return param(r.r, "vcs_provider_id")
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
	app.render("workspace_list.tmpl", w, r, struct {
		*otf.WorkspaceList
		organizationRoute
	}{
		WorkspaceList:     workspaces,
		organizationRoute: organizationRequest{r},
	})
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
	if err == otf.ErrResourceAlreadyExists {
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

	// Get existing perms as well as all teams in org
	perms, err := app.ListWorkspacePermissions(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	teams, err := app.ListTeams(r.Context(), *spec.OrganizationName)
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
		Roles          []otf.WorkspaceRole
	}{
		Workspace:      ws,
		LatestRun:      latest,
		LatestStreamID: "latest-" + otf.GenerateRandomString(5),
		Permissions:    perms,
		Teams:          teams,
		Roles: []otf.WorkspaceRole{
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
	if err := opts.Valid(); err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, editWorkspacePath(workspaceRequest{r}), http.StatusFound)
		return
	}
	// TODO: add support for updating vcs repo, e.g. branch, etc.
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
	providers, err := app.ListVCSProviders(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("workspace_vcs_provider_list.tmpl", w, r, struct {
		Items []*otf.VCSProvider
		workspaceRoute
	}{
		Items:          providers,
		workspaceRoute: workspaceRequest{r},
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

	provider, err := app.GetVCSProvider(r.Context(), opts.VCSProviderID, opts.OrganizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO(@leg100): how come this succeeds for gitlab when we're passing in a personal
	// access token and not an oauth token? On github, the two are the same
	// (AFAIK) so it makes sense that that works...
	client, err := provider.NewDirectoryClient(r.Context(), otf.DirectoryClientOptions{
		OAuthToken: &oauth2.Token{AccessToken: provider.Token()},
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repos, err := client.ListRepositories(r.Context(), opts.ListOptions)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("workspace_vcs_repo_list.tmpl", w, r, struct {
		*otf.RepoList
		vcsProviderRoute
	}{
		RepoList:         repos,
		vcsProviderRoute: vcsProviderRequest{workspaceRequest{r}},
	})
}

func (app *Application) connectWorkspaceRepo(w http.ResponseWriter, r *http.Request) {
	type options struct {
		otf.WorkspaceSpec
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Identifier    string `schema:"identifier,required"`
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.GetVCSProvider(r.Context(), opts.VCSProviderID, *opts.OrganizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client, err := provider.NewDirectoryClient(r.Context(), otf.DirectoryClientOptions{
		OAuthToken: &oauth2.Token{AccessToken: provider.Token()},
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo, err := client.GetRepository(r.Context(), opts.Identifier)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws, err := app.ConnectWorkspaceRepo(r.Context(), opts.WorkspaceSpec, otf.VCSRepo{
		Branch:     repo.Branch,
		HTTPURL:    repo.HTTPURL,
		Identifier: opts.Identifier,
		ProviderID: opts.VCSProviderID,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, getWorkspacePath(ws), http.StatusFound)
}

func (app *Application) disconnectWorkspaceRepo(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.All(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.DisconnectWorkspaceRepo(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, getWorkspacePath(ws), http.StatusFound)
}

func (app *Application) startRun(w http.ResponseWriter, r *http.Request) {
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

	run, err := startRun(r.Context(), app.Application, opts.WorkspaceSpec, speculative)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	http.Redirect(w, r, getRunPath(run), http.StatusFound)
}

func startRun(ctx context.Context, app otf.Application, spec otf.WorkspaceSpec, speculative bool) (*otf.Run, error) {
	ws, err := app.GetWorkspace(ctx, spec)
	if err != nil {
		return nil, err
	}

	var cv *otf.ConfigurationVersion
	opts := otf.ConfigurationVersionCreateOptions{
		Speculative: otf.Bool(speculative),
	}
	if ws.VCSRepo() != nil {
		provider, err := app.GetVCSProvider(ctx, ws.VCSRepo().ProviderID, ws.OrganizationName())
		if err != nil {
			return nil, err
		}
		client, err := provider.NewDirectoryClient(ctx, otf.DirectoryClientOptions{
			OAuthToken: &oauth2.Token{AccessToken: provider.Token()},
		})
		if err != nil {
			return nil, err
		}
		tarball, err := client.GetRepoTarball(ctx, ws.VCSRepo())
		if err != nil {
			return nil, err
		}
		cv, err = app.CreateConfigurationVersion(ctx, ws.ID(), opts)
		if err != nil {
			return nil, err
		}
		if err := app.UploadConfig(ctx, cv.ID(), tarball); err != nil {
			return nil, err
		}
	} else {
		latest, err := app.GetLatestConfigurationVersion(ctx, ws.ID())
		if err != nil {
			return nil, err
		}
		cv, err = app.CloneConfigurationVersion(ctx, latest.ID(), opts)
		if err != nil {
			return nil, err
		}
	}

	return app.CreateRun(ctx, spec, otf.RunCreateOptions{
		ConfigurationVersionID: otf.String(cv.ID()),
	})
}
