package workspace

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/rbac"
)

type web struct {
	otf.Renderer

	app application
}

func (h *web) addHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/workspaces", h.listWorkspaces).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/workspaces/new", h.newWorkspace).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/workspaces/create", h.createWorkspace).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", h.getWorkspaceByName).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}", h.getWorkspace).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/edit", h.editWorkspace).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/update", h.updateWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/delete", h.deleteWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/lock", h.lockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/unlock", h.unlockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/setup-connection-provider", h.listWorkspaceVCSProviders).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/setup-connection-repo", h.listWorkspaceVCSRepos).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/connect", h.connectWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/disconnect", h.disconnectWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/start-run", h.startRun).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/set-permission", h.setWorkspacePermission).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/unset-permission", h.unsetWorkspacePermission).Methods("POST")
}

func (h *web) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization    string `schema:"organization_name,required"`
		otf.ListOptions        // Pagination
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspaces, err := h.app.list(r.Context(), WorkspaceListOptions{
		Organization: &params.Organization,
		ListOptions:  params.ListOptions,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("workspace_list.tmpl", w, r, struct {
		*WorkspaceList
		Organization string
	}{
		WorkspaceList: workspaces,
		Organization:  params.Organization,
	})
}

func (h *web) newWorkspace(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("workspace_new.tmpl", w, r, organization)
}

func (h *web) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts otf.CreateWorkspaceOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.CreateWorkspace(r.Context(), opts)
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

func (h *web) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.GetWorkspace(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var latest *otf.Run
	if ws.LatestRunID() != nil {
		latest, err = h.GetRun(r.Context(), *ws.LatestRunID())
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.Render("workspace_get.tmpl", w, r, struct {
		*Workspace
		LatestRun      *otf.Run
		LatestStreamID string
	}{
		Workspace:      ws,
		LatestRun:      latest,
		LatestStreamID: "latest-" + otf.GenerateRandomString(5),
	})
}

func (h *web) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name         string `schema:"workspace_name,required"`
		Organization string `schema:"organization_name,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (h *web) editWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspace, err := h.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get existing perms as well as all teams in org
	perms, err := h.ListWorkspacePermissions(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	teams, err := h.ListTeams(r.Context(), workspace.Organization())
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

	h.Render("workspace_edit.tmpl", w, r, struct {
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

func (h *web) updateWorkspace(w http.ResponseWriter, r *http.Request) {
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
	ws, err := h.UpdateWorkspace(r.Context(), params.WorkspaceID, otf.UpdateWorkspaceOptions{
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

func (h *web) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace: "+ws.Name())
	http.Redirect(w, r, paths.Workspaces(ws.Organization()), http.StatusFound)
}

func (h *web) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.app.lock(r.Context(), id, nil)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (h *web) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.app.unlock(r.Context(), id, false)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (app *web) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
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

func (h *web) listWorkspaceVCSRepos(w http.ResponseWriter, r *http.Request) {
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

	ws, err := h.GetWorkspace(r.Context(), opts.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client, err := h.GetVCSClient(r.Context(), opts.VCSProviderID)
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

	h.Render("workspace_vcs_repo_list.tmpl", w, r, struct {
		Items []cloud.Repo
		*Workspace
		VCSProviderID string
	}{
		Items:         repos,
		Workspace:     ws,
		VCSProviderID: opts.VCSProviderID,
	})
}

func (h *web) connectWorkspace(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		otf.ConnectWorkspaceOptions
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.GetVCSProvider(r.Context(), params.ProviderID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params.Cloud = provider.CloudConfig().Name

	err = h.ConnectWorkspace(r.Context(), params.WorkspaceID, params.ConnectWorkspaceOptions)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}

func (h *web) disconnectWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.DisconnectWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
}

func (h *web) startRun(w http.ResponseWriter, r *http.Request) {
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

	ws, err := h.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	run, err := h.StartRun(r.Context(), params.WorkspaceID, otf.ConfigurationVersionCreateOptions{
		Speculative: otf.Bool(speculative),
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Workspace(ws.ID()), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID()), http.StatusFound)
}

func (h *web) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		TeamName    string `schema:"team_name,required"`
		Role        string `schema:"role,required"`
	}
	params := parameters{}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	role, err := rbac.WorkspaceRoleFromString(params.Role)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.SetWorkspacePermission(r.Context(), params.WorkspaceID, params.TeamName, role)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "updated workspace permissions")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID), http.StatusFound)
}

func (h *web) unsetWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		TeamName    string `schema:"team_name,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.UnsetWorkspacePermission(r.Context(), params.WorkspaceID, params.TeamName)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace permission")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID), http.StatusFound)
}
