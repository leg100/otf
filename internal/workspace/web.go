package workspace

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcsprovider"
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
		Search       string   `schema:"search[name],omitempty"`
		Tags         []string `schema:"search[tags],omitempty"`
		Organization *string  `schema:"organization_name,required"`
		PageNumber   int      `schema:"page[number]"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspaces, err := h.svc.ListWorkspaces(r.Context(), ListOptions{
		Search:       params.Search,
		Tags:         params.Tags,
		Organization: params.Organization,
		PageOptions: resource.PageOptions{
			PageNumber: params.PageNumber,
			PageSize:   html.PageSize,
		},
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retrieve all tags and create map, with each entry determining whether
	// listing is currently filtered by the tag or not.
	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Tag], error) {
		return h.svc.ListTags(r.Context(), *params.Organization, ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tagfilters := func() map[string]bool {
		m := make(map[string]bool, len(tags))
		for _, t := range tags {
			m[t.Name] = false
			for _, f := range params.Tags {
				if t.Name == f {
					m[t.Name] = true
					break
				}
			}
		}
		return m
	}

	response := struct {
		organization.OrganizationPage
		CreateWorkspaceAction rbac.Action
		*resource.Page[*Workspace]
		TagFilters map[string]bool
		Search     string
	}{
		OrganizationPage:      organization.NewPage(r, "workspaces", *params.Organization),
		CreateWorkspaceAction: rbac.CreateWorkspaceAction,
		Page:                  workspaces,
		TagFilters:            tagfilters(),
		Search:                params.Search,
	}

	if isHTMX := r.Header.Get("HX-Request"); isHTMX == "true" {
		if err := h.RenderTemplate("workspace_listing.tmpl", w, response); err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		h.Render("workspace_list.tmpl", w, response)
	}
}

func (h *webHandlers) newWorkspace(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
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
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.CreateWorkspace(r.Context(), CreateOptions{
		Name:         params.Name,
		Organization: params.Organization,
	})
	if err == internal.ErrResourceAlreadyExists {
		html.FlashError(w, "workspace already exists: "+*params.Name)
		http.Redirect(w, r, paths.NewWorkspace(*params.Organization), http.StatusFound)
		return
	}
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created workspace: "+ws.Name)
	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.GetWorkspace(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	policy, err := h.svc.GetPolicy(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := auth.UserFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var provider *vcsprovider.VCSProvider
	if ws.Connection != nil {
		provider, err = h.GetVCSProvider(r.Context(), ws.Connection.VCSProviderID)
		if err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Tag], error) {
		return h.svc.ListTags(r.Context(), ws.Organization, ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	getTagNames := func() (names []string) {
		for _, t := range tags {
			names = append(names, t.Name)
		}
		return
	}

	h.Render("workspace_get.tmpl", w, struct {
		WorkspacePage
		LockButton
		VCSProvider    *vcsprovider.VCSProvider
		CanApply       bool
		CanAddTags     bool
		CanRemoveTags  bool
		UnassignedTags []string
	}{
		WorkspacePage:  NewPage(r, ws.ID, ws),
		LockButton:     lockButtonHelper(ws, policy, user),
		VCSProvider:    provider,
		CanApply:       user.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		CanAddTags:     user.CanAccessWorkspace(rbac.AddTagsAction, policy),
		CanRemoveTags:  user.CanAccessWorkspace(rbac.RemoveTagsAction, policy),
		UnassignedTags: internal.DiffStrings(getTagNames(), ws.Tags),
	})
}

func (h *webHandlers) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string `schema:"workspace_name,required"`
		Organization string `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) editWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspace, err := h.svc.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	policy, err := h.svc.GetPolicy(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get teams that have yet to be assigned a permission
	teams, err := h.ListTeams(r.Context(), workspace.Organization)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var provider *vcsprovider.VCSProvider
	if workspace.Connection != nil {
		provider, err = h.GetVCSProvider(r.Context(), workspace.Connection.VCSProviderID)
		if err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Tag], error) {
		return h.svc.ListTags(r.Context(), workspace.Organization, ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	getTagNames := func() (names []string) {
		for _, t := range tags {
			names = append(names, t.Name)
		}
		return
	}

	h.Render("workspace_edit.tmpl", w, struct {
		WorkspacePage
		Policy                         internal.WorkspacePolicy
		Unassigned                     []*auth.Team
		Roles                          []rbac.Role
		VCSProvider                    *vcsprovider.VCSProvider
		UnassignedTags                 []string
		UpdateWorkspaceAction          rbac.Action
		DeleteWorkspaceAction          rbac.Action
		SetWorkspacePermissionAction   rbac.Action
		UnsetWorkspacePermissionAction rbac.Action
		AddTagsAction                  rbac.Action
		RemoveTagsAction               rbac.Action
		CreateRunAction                rbac.Action
	}{
		WorkspacePage: NewPage(r, "edit | "+workspace.ID, workspace),
		Policy:        policy,
		Unassigned:    filterUnassigned(policy, teams),
		Roles: []rbac.Role{
			rbac.WorkspaceReadRole,
			rbac.WorkspacePlanRole,
			rbac.WorkspaceWriteRole,
			rbac.WorkspaceAdminRole,
		},
		VCSProvider:                    provider,
		UnassignedTags:                 internal.DiffStrings(getTagNames(), workspace.Tags),
		UpdateWorkspaceAction:          rbac.UpdateWorkspaceAction,
		DeleteWorkspaceAction:          rbac.DeleteWorkspaceAction,
		SetWorkspacePermissionAction:   rbac.SetWorkspacePermissionAction,
		UnsetWorkspacePermissionAction: rbac.UnsetWorkspacePermissionAction,
		CreateRunAction:                rbac.CreateRunAction,
		AddTagsAction:                  rbac.AddTagsAction,
		RemoveTagsAction:               rbac.RemoveTagsAction,
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
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
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
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, paths.EditWorkspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace: "+ws.Name)
	http.Redirect(w, r, paths.Workspaces(ws.Organization), http.StatusFound)
}

func (h *webHandlers) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.LockWorkspace(r.Context(), id, nil)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.UnlockWorkspace(r.Context(), workspaceID, nil, false)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.UnlockWorkspace(r.Context(), workspaceID, nil, true)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	providers, err := h.ListVCSProviders(r.Context(), ws.Organization)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
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
		WorkspaceID   string `schema:"workspace_id,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`

		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.svc.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client, err := h.GetVCSClient(r.Context(), params.VCSProviderID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repos, err := client.ListRepositories(r.Context(), cloud.ListRepositoriesOptions{
		PageSize: html.PageSize,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
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
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err := h.svc.connect(r.Context(), params.WorkspaceID, params.ConnectOptions)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) disconnect(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.svc.disconnect(r.Context(), workspaceID); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, paths.Workspace(workspaceID), http.StatusFound)
}

func (h *webHandlers) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		TeamName    string `schema:"team_name,required"`
		Role        string `schema:"role,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	role, err := rbac.WorkspaceRoleFromString(params.Role)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.svc.SetPermission(r.Context(), params.WorkspaceID, params.TeamName, role)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
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
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.svc.UnsetPermission(r.Context(), params.WorkspaceID, params.TeamName)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace permission")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID), http.StatusFound)
}

// filterUnassigned removes from the list of teams those that are part of the
// policy, i.e. those that have been assigned a permission.
//
// NOTE: the owners team is always removed because by default it is assigned the
// admin role.
func filterUnassigned(policy internal.WorkspacePolicy, teams []*auth.Team) (unassigned []*auth.Team) {
	assigned := make(map[string]struct{}, len(teams))
	for _, p := range policy.Permissions {
		assigned[p.Team] = struct{}{}
	}
	for _, t := range teams {
		if t.Name == "owners" {
			continue
		}
		if _, ok := assigned[t.Name]; !ok {
			unassigned = append(unassigned, t)
		}
	}
	return
}
