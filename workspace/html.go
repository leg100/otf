package workspace

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/rbac"
	"github.com/r3labs/sse/v2"
)

type htmlApp struct {
	otf.Renderer

	app service
}

func (app *htmlApp) AddHTMLHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/workspaces", app.listWorkspaces).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/workspaces/new", app.newWorkspace).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/workspaces/create", app.createWorkspace).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", app.getWorkspaceByName).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}", app.getWorkspace).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/edit", app.editWorkspace).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/update", app.updateWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/delete", app.deleteWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/lock", app.lockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/unlock", app.unlockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/set-permission", app.setWorkspacePermission).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/unset-permission", app.unsetWorkspacePermission).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/setup-connection-provider", app.listWorkspaceVCSProviders).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/setup-connection-repo", app.listWorkspaceVCSRepos).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/connect", app.connectWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/disconnect", app.disconnectWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/start-run", app.startRun).Methods("POST")
}

func (app *htmlApp) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization    string `schema:"organization_name,required"`
		otf.ListOptions        // Pagination
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspaces, err := app.ListWorkspaces(r.Context(), otf.WorkspaceListOptions{
		Organization: &params.Organization,
		ListOptions:  params.ListOptions,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("workspace_list.tmpl", w, r, struct {
		*otf.WorkspaceList
		Organization string
	}{
		WorkspaceList: workspaces,
		Organization:  params.Organization,
	})
}

func (app *htmlApp) newWorkspace(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	app.Render("workspace_new.tmpl", w, r, organization)
}

func (app *htmlApp) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts otf.CreateWorkspaceOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.CreateWorkspace(r.Context(), opts)
	if err == otf.ErrResourceAlreadyExists {
		html.FlashError(w, "workspace already exists: "+*opts.Name)
		http.Redirect(w, r, paths.NewWorkspace(*opts.Organization), http.StatusFound)
		return
	}
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created workspace: "+ws.Name())
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *htmlApp) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var latest *otf.Run
	if ws.LatestRunID() != nil {
		latest, err = app.GetRun(r.Context(), *ws.LatestRunID())
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	app.Render("workspace_get.tmpl", w, r, struct {
		*Workspace
		LatestRun      *otf.Run
		LatestStreamID string
	}{
		Workspace:      ws,
		LatestRun:      latest,
		LatestStreamID: "latest-" + otf.GenerateRandomString(5),
	})
}

func (app *htmlApp) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name         string `schema:"workspace_name,required"`
		Organization string `schema:"organization_name,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *htmlApp) editWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspace, err := app.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get existing perms as well as all teams in org
	perms, err := app.ListWorkspacePermissions(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	teams, err := app.ListTeams(r.Context(), workspace.Organization())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
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

	app.Render("workspace_edit.tmpl", w, r, struct {
		*Workspace
		Permissions []*otf.WorkspacePermission
		Teams       []*otf.Team
		Roles       []rbac.Role
	}{
		Workspace:   workspace,
		Permissions: perms,
		Teams:       teams,
		Roles: []rbac.Role{
			rbac.WorkspaceReadRole,
			rbac.WorkspacePlanRole,
			rbac.WorkspaceWriteRole,
			rbac.WorkspaceAdminRole,
		},
	})
}

func (app *htmlApp) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		AutoApply        bool `schema:"auto_apply"`
		Name             *string
		Description      *string
		ExecutionMode    *otf.ExecutionMode `schema:"execution_mode"`
		TerraformVersion *string            `schema:"terraform_version"`
		WorkingDirectory *string            `schema:"working_directory"`
		WorkspaceID      string             `schema:"workspace_id,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// TODO: add support for updating vcs repo, e.g. branch, etc.
	ws, err := app.UpdateWorkspace(r.Context(), params.WorkspaceID, otf.UpdateWorkspaceOptions{
		AutoApply:        &params.AutoApply,
		Name:             params.Name,
		Description:      params.Description,
		ExecutionMode:    params.ExecutionMode,
		TerraformVersion: params.TerraformVersion,
		WorkingDirectory: params.WorkingDirectory,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, paths.EditWorkspace(ws.ID()), http.StatusFound)
}

func (app *htmlApp) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace: "+ws.Name())
	http.Redirect(w, r, paths.Workspaces(ws.Organization()), http.StatusFound)
}

func (app *htmlApp) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.LockWorkspace(r.Context(), id, otf.WorkspaceLockOptions{})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *htmlApp) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.UnlockWorkspace(r.Context(), id, otf.WorkspaceUnlockOptions{})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *htmlApp) watchWorkspace(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		StreamID    string `schema:"stream,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	events, err := app.Watch(r.Context(), otf.WatchOptions{
		WorkspaceID: otf.String(ws.ID()),
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
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

func (app *htmlApp) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	providers, err := app.ListVCSProviders(r.Context(), ws.Organization())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("workspace_vcs_provider_list.tmpl", w, r, struct {
		Items []otf.VCSProvider
		*Workspace
	}{
		Items:     providers,
		Workspace: ws,
	})
}

func (app *htmlApp) listWorkspaceVCSRepos(w http.ResponseWriter, r *http.Request) {
	type options struct {
		WorkspaceID     string `schema:"workspace_id,required"`
		VCSProviderID   string `schema:"vcs_provider_id,required"`
		otf.ListOptions        // Pagination
		// TODO: filters, public/private, etc
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), opts.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client, err := app.GetVCSClient(r.Context(), opts.VCSProviderID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repos, err := client.ListRepositories(r.Context(), cloud.ListRepositoriesOptions{
		PageSize: 100,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("workspace_vcs_repo_list.tmpl", w, r, struct {
		Items []cloud.Repo
		*Workspace
		VCSProviderID string
	}{
		Items:         repos,
		Workspace:     ws,
		VCSProviderID: opts.VCSProviderID,
	})
}

func (app *htmlApp) connectWorkspace(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		otf.ConnectWorkspaceOptions
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.GetVCSProvider(r.Context(), params.ProviderID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params.Cloud = provider.CloudConfig().Name

	err = app.ConnectWorkspace(r.Context(), params.WorkspaceID, params.ConnectWorkspaceOptions)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}

func (app *htmlApp) disconnectWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.DisconnectWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *htmlApp) startRun(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		Strategy    string `schema:"strategy,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var speculative bool
	switch params.Strategy {
	case "plan-only":
		speculative = true
	case "plan-and-apply":
		speculative = false
	default:
		html.Error(w, "invalid strategy", http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	run, err := app.StartRun(r.Context(), params.WorkspaceID, otf.ConfigurationVersionCreateOptions{
		Speculative: otf.Bool(speculative),
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID()), http.StatusFound)
}
