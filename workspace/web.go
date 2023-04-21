package workspace

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/vcsprovider"
)

type (
	webHandlers struct {
		html.Renderer
		auth.TeamService
		VCSProviderService

		svc Service
	}

	// WorkspacePage contains data shared by all workspace-based pages.
	WorkspacePage struct {
		organization.OrganizationPage

		Workspace *Workspace
	}
)

func NewPage(r *http.Request, title string, workspace *Workspace) WorkspacePage {
	return WorkspacePage{
		OrganizationPage: organization.NewPage(r, title, workspace.Organization),
		Workspace:        workspace,
	}
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

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
	r.HandleFunc("/workspaces/{workspace_id}/force-unlock", h.forceUnlockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/setup-connection-provider", h.listWorkspaceVCSProviders).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/setup-connection-repo", h.listWorkspaceVCSRepos).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/connect", h.connect).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/disconnect", h.disconnect).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/set-permission", h.setWorkspacePermission).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/unset-permission", h.unsetWorkspacePermission).Methods("POST")
}

func (h *webHandlers) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization    string `schema:"organization_name,required"`
		otf.ListOptions        // Pagination
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspaces, err := h.svc.ListWorkspaces(r.Context(), ListOptions{
		Organization: &params.Organization,
		ListOptions:  params.ListOptions,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("workspace_list.tmpl", w, struct {
		organization.OrganizationPage
		*WorkspaceList
	}{
		OrganizationPage: organization.NewPage(r, "workspaces", params.Organization),
		WorkspaceList:    workspaces,
	})
}

func (h *webHandlers) newWorkspace(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("workspace_new.tmpl", w, struct {
		organization.OrganizationPage
	}{
		OrganizationPage: organization.NewPage(r, "new workspace", org),
	})
}

func (h *webHandlers) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         *string `schema:"name,required"`
		Organization *string `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.CreateWorkspace(r.Context(), CreateOptions{
		Name:         params.Name,
		Organization: params.Organization,
	})
	if err == otf.ErrResourceAlreadyExists {
		html.FlashError(w, "workspace already exists: "+*params.Name)
		http.Redirect(w, r, paths.NewWorkspace(*params.Organization), http.StatusFound)
		return
	}
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created workspace: "+ws.Name)
	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.GetWorkspace(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	policy, err := h.svc.GetPolicy(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := auth.UserFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("workspace_get.tmpl", w, struct {
		WorkspacePage
		LockButton
	}{
		WorkspacePage: NewPage(r, ws.ID, ws),
		LockButton:    lockButtonHelper(ws, policy, user),
	})
}

func (h *webHandlers) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string `schema:"workspace_name,required"`
		Organization string `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) editWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspace, err := h.svc.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	policy, err := h.svc.GetPolicy(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get teams that have yet to be assigned a permission
	unassigned, err := h.ListTeams(r.Context(), workspace.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, perm := range policy.Permissions {
		for it, t := range unassigned {
			if t.ID == perm.Team {
				unassigned = append(unassigned[:it], unassigned[it+1:]...)
				break
			}
		}
	}

	h.Render("workspace_edit.tmpl", w, struct {
		WorkspacePage
		Permissions []otf.WorkspacePermission
		Unassigned  []*auth.Team
		Roles       []rbac.Role
	}{
		WorkspacePage: NewPage(r, "edit | "+workspace.ID, workspace),
		Permissions:   policy.Permissions,
		Unassigned:    unassigned,
		Roles: []rbac.Role{
			rbac.WorkspaceReadRole,
			rbac.WorkspacePlanRole,
			rbac.WorkspaceWriteRole,
			rbac.WorkspaceAdminRole,
		},
	})
}

func (h *webHandlers) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		AutoApply        bool `schema:"auto_apply"`
		Name             *string
		Description      *string
		ExecutionMode    *ExecutionMode `schema:"execution_mode"`
		TerraformVersion *string        `schema:"terraform_version"`
		WorkingDirectory *string        `schema:"working_directory"`
		WorkspaceID      string         `schema:"workspace_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// TODO: add support for updating vcs repo, e.g. branch, etc.
	ws, err := h.svc.UpdateWorkspace(r.Context(), params.WorkspaceID, UpdateOptions{
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
	http.Redirect(w, r, paths.EditWorkspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace: "+ws.Name)
	http.Redirect(w, r, paths.Workspaces(ws.Organization), http.StatusFound)
}

func (h *webHandlers) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.LockWorkspace(r.Context(), id, nil)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.UnlockWorkspace(r.Context(), workspaceID, nil, false)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.UnlockWorkspace(r.Context(), workspaceID, nil, true)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	providers, err := h.ListVCSProviders(r.Context(), ws.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("workspace_vcs_provider_list.tmpl", w, struct {
		WorkspacePage
		Items []*vcsprovider.VCSProvider
	}{
		WorkspacePage: NewPage(r, "list vcs providers | "+ws.ID, ws),
		Items:         providers,
	})
}

func (h *webHandlers) listWorkspaceVCSRepos(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID     string `schema:"workspace_id,required"`
		VCSProviderID   string `schema:"vcs_provider_id,required"`
		otf.ListOptions        // Pagination
		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client, err := h.GetVCSClient(r.Context(), params.VCSProviderID)
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

	h.Render("workspace_vcs_repo_list.tmpl", w, struct {
		WorkspacePage
		Repos         []string
		VCSProviderID string
	}{
		WorkspacePage: NewPage(r, "list vcs repos | "+ws.ID, ws),
		Repos:         repos,
		VCSProviderID: params.VCSProviderID,
	})
}

func (h *webHandlers) connect(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		ConnectOptions
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err := h.svc.connect(r.Context(), params.WorkspaceID, params.ConnectOptions)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) disconnect(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var stack html.FlashStack
	err = h.svc.disconnect(r.Context(), workspaceID)
	if errors.Is(err, otf.ErrWarning) {
		stack.Push(html.FlashWarningType, err.Error())
	} else if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stack.Push(html.FlashSuccessType, "disconnected workspace from repo")
	stack.Write(w)

	http.Redirect(w, r, paths.Workspace(workspaceID), http.StatusFound)
}

func (h *webHandlers) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		TeamName    string `schema:"team_name,required"`
		Role        string `schema:"role,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	role, err := rbac.WorkspaceRoleFromString(params.Role)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.svc.SetPermission(r.Context(), params.WorkspaceID, params.TeamName, role)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "updated workspace permissions")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) unsetWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		TeamName    string `schema:"team_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.svc.UnsetPermission(r.Context(), params.WorkspaceID, params.TeamName)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace permission")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID), http.StatusFound)
}
