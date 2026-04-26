package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	enginepkg "github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	runpkg "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sshkey"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/path"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

type Handlers struct {
	Client         WorkspaceService
	Authorizer     authz.Interface
	SingleRunTable func(*runpkg.Run) templ.Component
}

type WorkspaceService interface {
	GetWorkspace(context.Context, resource.TfeID) (*workspace.Workspace, error)
	ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)
	ListTags(ctx context.Context, organization organization.Name, opts workspace.ListTagsOptions) (*resource.Page[*workspace.Tag], error)
	CreateWorkspace(ctx context.Context, opts workspace.CreateOptions) (*workspace.Workspace, error)
	GetWorkspaceByName(ctx context.Context, organization organization.Name, name string) (*workspace.Workspace, error)
	GetWorkspacePolicy(ctx context.Context, workspaceID resource.TfeID) (workspace.Policy, error)
	UpdateWorkspace(ctx context.Context, workspaceID resource.TfeID, opts workspace.UpdateOptions) (*workspace.Workspace, error)
	DeleteWorkspace(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
	Lock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) (*workspace.Workspace, error)
	Unlock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID, force bool) (*workspace.Workspace, error)
	SetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.TfeID, role authz.Role) error
	UnsetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.TfeID) error
	DeleteTags(ctx context.Context, organization organization.Name, tagIDs []resource.TfeID) error
	AddTags(ctx context.Context, workspaceID resource.TfeID, tags []workspace.TagSpec) error
	RemoveTags(ctx context.Context, workspaceID resource.TfeID, tags []workspace.TagSpec) error
	ListTeams(ctx context.Context, organization organization.Name) ([]*team.Team, error)
	GetVCSProvider(ctx context.Context, id resource.TfeID) (*vcs.Provider, error)
	ListVCSProviders(ctx context.Context, organization organization.Name) ([]*vcs.Provider, error)
	GetLatest(ctx context.Context, e *enginepkg.Engine) (string, time.Time, error)
	ListSSHKeys(ctx context.Context, org organization.Name) ([]*sshkey.SSHKey, error)
	GetRun(ctx context.Context, id resource.TfeID) (*runpkg.Run, error)
}

func (h *Handlers) AddHandlers(r *mux.Router) {
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

	r.HandleFunc("/workspaces/{workspace_id}/edit-permissions", h.editPermissions).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/set-permission", h.setWorkspacePermission).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/unset-permission", h.unsetWorkspacePermission).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/edit-engine", h.editEngine).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/update-engine", h.updateEngine).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/edit-vcs", h.editVCS).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/update-vcs", h.updateVCS).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/setup-connection-provider", h.listWorkspaceVCSProviders).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/setup-connection-repo", h.listWorkspaceVCSRepos).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/connect", h.connect).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/disconnect", h.disconnect).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/create-tag", h.createTag).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/delete-tag", h.deleteTag).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/edit-ssh-key", h.editWorkspaceSSHKey).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/update-ssh-key", h.updateWorkspaceSSHKey).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/edit-advanced", h.editAdvanced).Methods("GET")
}

func (h *Handlers) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var params struct {
		workspace.ListOptions
		StatusFilterVisible bool `schema:"status_filter_visible"`
		TagFilterVisible    bool `schema:"tag_filter_visible"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// retrieve all tags for listing in tag filter
	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Tag], error) {
		return h.Client.ListTags(r.Context(), *params.Organization, workspace.ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	tagStrings := make([]string, len(tags))
	for i, tag := range tags {
		tagStrings[i] = tag.Name
	}

	page, err := h.Client.ListWorkspaces(r.Context(), params.ListOptions)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := workspaceListProps{
		organization:        *params.Organization,
		allTags:             tagStrings,
		selectedTags:        params.Tags,
		search:              params.Search,
		status:              params.Status,
		statusFilterVisible: params.StatusFilterVisible,
		tagFilterVisible:    params.TagFilterVisible,
		pageOptions:         params.PageOptions,
		page:                page,
	}

	helpers.RenderPage(
		workspaceList(props),
		"workspaces",
		w,
		r,
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Workspaces"},
		),
		helpers.WithOrganization(*params.Organization),
		helpers.WithContentActions(workspaceListActions(
			*params.Organization,
			h.Authorizer.CanAccess(r.Context(), authz.CreateWorkspaceAction, *params.Organization),
		)),
	)
}

func (h *Handlers) newWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	helpers.RenderPage(
		workspaceNew(params.Organization),
		"new workspace",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "workspaces", Link: path.List(resource.WorkspaceKind, params.Organization)},
			helpers.Breadcrumb{Name: "new"},
		),
	)
}

func (h *Handlers) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         *string            `schema:"name,required"`
		Organization *organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.CreateWorkspace(r.Context(), workspace.CreateOptions{
		Name:         params.Name,
		Organization: params.Organization,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "created workspace: "+ws.Name)
	http.Redirect(w, r, path.Get(ws.ID), http.StatusFound)
}

func (h *Handlers) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	user, err := user.UserFromContext(r.Context())
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	var provider *vcs.Provider
	if ws.Connection != nil {
		provider, err = h.Client.GetVCSProvider(r.Context(), ws.Connection.VCSProviderID)
		if err != nil {
			helpers.Error(r, w, err.Error())
			return
		}
	}

	// retrieve tags that are available to be assigned to the workspace
	// (excluding those already assigned to the workspace).
	var availableTags []string
	{
		tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Tag], error) {
			return h.Client.ListTags(r.Context(), ws.Organization, workspace.ListTagsOptions{
				PageOptions: opts,
			})
		})
		if err != nil {
			helpers.Error(r, w, err.Error())
			return
		}
		names := internal.Map(tags, func(t *workspace.Tag) string {
			return t.Name
		})
		availableTags = internal.Diff(names, ws.Tags)
	}

	lockInfo, err := h.lockButtonHelper(r.Context(), ws, user)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// Generate component for latest run if workspace has one.
	var latestRunTable templ.Component
	if ws.LatestRun != nil {
		run, err := h.Client.GetRun(r.Context(), ws.LatestRun.ID)
		if err != nil {
			helpers.Error(r, w, err.Error())
			return
		}
		latestRunTable = h.SingleRunTable(run)
	}

	props := workspaceGetProps{
		ws:                 ws,
		workspaceLockInfo:  lockInfo,
		vcsProvider:        provider,
		canApply:           h.Authorizer.CanAccess(r.Context(), authz.ApplyRunAction, ws.ID),
		canAddTags:         h.Authorizer.CanAccess(r.Context(), authz.AddTagsAction, ws.ID),
		canRemoveTags:      h.Authorizer.CanAccess(r.Context(), authz.RemoveTagsAction, ws.ID),
		canCreateRun:       h.Authorizer.CanAccess(r.Context(), authz.CreateRunAction, ws.ID),
		canLockWorkspace:   h.Authorizer.CanAccess(r.Context(), authz.LockWorkspaceAction, ws.ID),
		canUnlockWorkspace: h.Authorizer.CanAccess(r.Context(), authz.UnlockWorkspaceAction, ws.ID),
		canUpdateWorkspace: h.Authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, ws.ID),
		tagsDropdown: helpers.SearchDropdownProps{
			Name:        "tag_name",
			Available:   availableTags,
			Existing:    ws.Tags,
			Action:      templ.SafeURL(path.Resource(resource.Action("create-tag"), ws.ID)),
			Placeholder: "Add tags",
			Width:       helpers.NarrowDropDown,
		},
		latestRunTable: latestRunTable,
	}
	helpers.RenderPage(
		workspaceGet(props),
		"workspaces",
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: props.ws.Name},
		),
	)
}

func (h *Handlers) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string            `schema:"workspace_name,required"`
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, path.Get(ws.ID), http.StatusFound)
}

func (h *Handlers) editWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	poolsURL := path.Resource(resource.Action("pools"), workspaceID)
	if ws.AgentPoolID != nil {
		poolsURL += "?agent_pool_id=" + ws.AgentPoolID.String()
	}

	props := workspaceEditProps{
		ws:       ws,
		poolsURL: poolsURL,
	}
	helpers.RenderPage(
		workspaceEdit(props),
		"edit | "+ws.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithSideMenu(helpers.WorkspaceSettingsMenu(ws.ID)),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "General Settings"},
		),
	)
}

func (h *Handlers) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		AgentPoolID           *resource.TfeID `schema:"agent_pool_id"`
		AutoApply             bool            `schema:"auto_apply"`
		Name                  string
		Description           string
		ExecutionMode         workspace.ExecutionMode `schema:"execution_mode"`
		Engine                *enginepkg.Engine       `schema:"engine"`
		LatestEngineVersion   bool                    `schema:"latest_engine_version"`
		SpecificEngineVersion *workspace.Version      `schema:"specific_engine_version"`
		WorkingDirectory      string                  `schema:"working_directory"`
		WorkspaceID           resource.TfeID          `schema:"workspace_id,required"`
		GlobalRemoteState     bool                    `schema:"global_remote_state"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	opts := workspace.UpdateOptions{
		AutoApply:         &params.AutoApply,
		Name:              &params.Name,
		Description:       &params.Description,
		ExecutionMode:     &params.ExecutionMode,
		Engine:            params.Engine,
		WorkingDirectory:  &params.WorkingDirectory,
		GlobalRemoteState: &params.GlobalRemoteState,
	}
	if params.LatestEngineVersion {
		opts.EngineVersion = &workspace.Version{Latest: true}
	} else {
		opts.EngineVersion = params.SpecificEngineVersion
	}
	// only set agent pool ID if execution mode is set to agent
	if params.ExecutionMode == workspace.AgentExecutionMode {
		opts.AgentPoolID = params.AgentPoolID
	}

	ws, err := h.Client.UpdateWorkspace(r.Context(), params.WorkspaceID, opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, path.Edit(ws.ID), http.StatusFound)
}

func (h *Handlers) editWorkspaceSSHKey(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	keys, err := h.Client.ListSSHKeys(r.Context(), ws.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := workspaceEditSSHKeyProps{
		ws:   ws,
		keys: keys,
	}
	helpers.RenderPage(
		workspaceEditSSHKey(props),
		"workspace ssh key | "+ws.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithSideMenu(helpers.WorkspaceSettingsMenu(ws.ID)),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "SSH Key"},
		),
	)
}

func (h *Handlers) updateWorkspaceSSHKey(w http.ResponseWriter, r *http.Request) {
	var params struct {
		SSHKeyID    resource.TfeID `schema:"ssh_key_id"`
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	opts := &workspace.UpdateSSHKeyOptions{}
	if !params.SSHKeyID.IsZero() {
		opts.SSHKeyID = &params.SSHKeyID
	}

	ws, err := h.Client.UpdateWorkspace(r.Context(), params.WorkspaceID, workspace.UpdateOptions{
		UpdateSSHKeyOptions: opts,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, path.Resource(resource.Action("edit-ssh-key"), ws.ID), http.StatusFound)
}

func (h *Handlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "deleted workspace: "+ws.Name)
	http.Redirect(w, r, path.List(resource.WorkspaceKind, ws.Organization), http.StatusFound)
}

func (h *Handlers) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.Lock(r.Context(), id, nil)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, path.Get(ws.ID), http.StatusFound)
}

func (h *Handlers) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.Unlock(r.Context(), workspaceID, nil, false)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, path.Get(ws.ID), http.StatusFound)
}

func (h *Handlers) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.Unlock(r.Context(), workspaceID, nil, true)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, path.Get(ws.ID), http.StatusFound)
}

func (h *Handlers) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	providers, err := h.Client.ListVCSProviders(r.Context(), ws.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.RenderPage(
		listVCSProviders(ws, providers),
		"list vcs providers | "+ws.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Connect workspace to VCS provider"},
		),
	)
}

func (h *Handlers) listWorkspaceVCSRepos(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID   resource.TfeID `schema:"workspace_id,required"`
		VCSProviderID resource.TfeID `schema:"vcs_provider_id,required"`

		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	client, err := h.Client.GetVCSProvider(r.Context(), params.VCSProviderID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	repos, err := client.ListRepositories(r.Context(), vcs.ListRepositoriesOptions{
		PageSize: resource.MaxPageSize,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.RenderPage(
		listVCSRepos(ws, params.VCSProviderID, repos),
		"list vcs repos | "+ws.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Connect workspace to VCS repository"},
		),
	)
}

func (h *Handlers) connect(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID   resource.TfeID  `schema:"workspace_id,required"`
		RepoPath      *vcs.Repo       `schema:"identifier,required"`
		VCSProviderID *resource.TfeID `schema:"vcs_provider_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	_, err := h.Client.UpdateWorkspace(r.Context(), params.WorkspaceID, workspace.UpdateOptions{
		ConnectOptions: &workspace.ConnectOptions{
			VCSProviderID: params.VCSProviderID,
			RepoPath:      params.RepoPath,
		},
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, path.Resource(resource.Action("edit-vcs"), params.WorkspaceID), http.StatusFound)
}

func (h *Handlers) disconnect(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	_, err = h.Client.UpdateWorkspace(r.Context(), workspaceID, workspace.UpdateOptions{
		Disconnect: true,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, path.Get(workspaceID), http.StatusFound)
}

func (h *Handlers) editPermissions(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	policy, err := h.Client.GetWorkspacePolicy(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// Get teams for populating team permissions
	teams, err := h.Client.ListTeams(r.Context(), ws.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// want current policy permissions to include not only team ID but team name
	// too for user's benefit
	perms := make([]workspacePerm, len(policy.Permissions))
	for i, pp := range policy.Permissions {
		// get team name corresponding to team ID
		for _, t := range teams {
			if t.ID == pp.TeamID {
				perms[i] = workspacePerm{
					role: pp.Role,
					team: t,
				}
				break
			}
		}
	}

	props := editPermissionsProps{
		ws:         ws,
		assigned:   perms,
		unassigned: filterUnassigned(policy, teams),
		roles: []authz.Role{
			authz.WorkspaceReadRole,
			authz.WorkspacePlanRole,
			authz.WorkspaceWriteRole,
			authz.WorkspaceAdminRole,
		},
	}
	helpers.RenderPage(
		editPermissions(props),
		"edit permissions | "+ws.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithSideMenu(helpers.WorkspaceSettingsMenu(ws.ID)),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Permissions"},
		),
	)
}

func (h *Handlers) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		TeamID      resource.TfeID `schema:"team_id,required"`
		Role        string         `schema:"role,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	role, err := authz.WorkspaceRoleFromString(params.Role)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	err = h.Client.SetWorkspacePermission(r.Context(), params.WorkspaceID, params.TeamID, role)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "updated workspace permissions")
	http.Redirect(w, r, path.Resource(resource.Action("edit-permissions"), params.WorkspaceID), http.StatusFound)
}

func (h *Handlers) unsetWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		TeamID      resource.TfeID `schema:"team_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.Client.UnsetWorkspacePermission(r.Context(), params.WorkspaceID, params.TeamID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "deleted workspace permission")
	http.Redirect(w, r, path.Resource(resource.Action("edit-permissions"), params.WorkspaceID), http.StatusFound)
}

type engine struct {
	name    string
	latest  bool
	version string
}

func (h *Handlers) editEngine(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// Construct list of engines for template
	engines := make([]engine, len(enginepkg.Engines()))
	current := ""
	for i, engine := range enginepkg.Engines() {
		engines[i].name = engine.String()
		if engine.String() == ws.Engine.String() {
			current = engine.String()
			engines[i].latest = ws.EngineVersion.Latest
		}
		// Offer the user the latest available version for an engine if:
		// (a): it's not the current engine, or
		// (b): it's currently set to track the latest version.
		if current != engine.String() || engines[i].latest {
			latest, _, err := h.Client.GetLatest(r.Context(), engine)
			if err != nil {
				helpers.Error(r, w, err.Error())
				return
			}
			engines[i].version = latest
		} else {
			engines[i].version = ws.EngineVersion.String()
		}
	}

	props := editEngineProps{
		ws:      ws,
		engines: engines,
		current: current,
	}
	helpers.RenderPage(
		editEngine(props),
		"edit engine | "+ws.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithSideMenu(helpers.WorkspaceSettingsMenu(ws.ID)),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Engines"},
		),
	)
}

func (h *Handlers) updateEngine(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID           resource.TfeID     `schema:"workspace_id,required"`
		Engine                *enginepkg.Engine  `schema:"engine"`
		LatestEngineVersion   bool               `schema:"latest_engine_version"`
		SpecificEngineVersion *workspace.Version `schema:"specific_engine_version"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	opts := workspace.UpdateOptions{
		Engine: params.Engine,
	}
	if params.LatestEngineVersion {
		opts.EngineVersion = &workspace.Version{Latest: true}
	} else {
		opts.EngineVersion = params.SpecificEngineVersion
	}
	ws, err := h.Client.UpdateWorkspace(r.Context(), params.WorkspaceID, opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "updated engine")
	http.Redirect(w, r, path.Resource(resource.Action("edit-engine"), ws.ID), http.StatusFound)
}

func (h *Handlers) editVCS(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	var provider *vcs.Provider
	if ws.Connection != nil {
		provider, err = h.Client.GetVCSProvider(r.Context(), ws.Connection.VCSProviderID)
		if err != nil {
			helpers.Error(r, w, err.Error())
			return
		}
	}

	props := editVCSProps{
		ws:                 ws,
		vcsProvider:        provider,
		canUpdateWorkspace: h.Authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, ws.ID),
		canDeleteWorkspace: h.Authorizer.CanAccess(r.Context(), authz.DeleteWorkspaceAction, ws.ID),
	}
	helpers.RenderPage(
		editVCS(props),
		"edit vcs | "+ws.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithSideMenu(helpers.WorkspaceSettingsMenu(ws.ID)),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "VCS Settings"},
		),
	)
}

func (h *Handlers) updateVCS(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID         resource.TfeID `schema:"workspace_id,required"`
		VCSTriggerStrategy  string         `schema:"vcs_trigger"`
		TriggerPatternsJSON string         `schema:"trigger_patterns"`
		VCSBranch           string         `schema:"vcs_branch"`
		PredefinedTagsRegex string         `schema:"tags_regex"`
		CustomTagsRegex     string         `schema:"custom_tags_regex"`
		AllowCLIApply       bool           `schema:"allow_cli_apply"`
		SpeculativeEnabled  bool           `schema:"speculative_enabled"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// get workspace before updating to determine if it is connected or not.
	ws, err := h.Client.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	opts := workspace.UpdateOptions{
		SpeculativeEnabled: &params.SpeculativeEnabled,
	}

	if ws.Connection != nil {
		// workspace is connected, so set connection fields
		opts.ConnectOptions = &workspace.ConnectOptions{
			AllowCLIApply: &params.AllowCLIApply,
			Branch:        &params.VCSBranch,
		}
		switch params.VCSTriggerStrategy {
		case vcsTriggerAlways:
			opts.AlwaysTrigger = new(true)
		case vcsTriggerPatterns:
			err := json.Unmarshal([]byte(params.TriggerPatternsJSON), &opts.TriggerPatterns)
			if err != nil {
				helpers.Error(r, w, err.Error())
				return
			}
		case vcsTriggerTags:
			if params.PredefinedTagsRegex == vcsTagRegexCustom {
				opts.ConnectOptions.TagsRegex = &params.CustomTagsRegex
			} else {
				opts.ConnectOptions.TagsRegex = &params.PredefinedTagsRegex
			}
		}
	}

	ws, err = h.Client.UpdateWorkspace(r.Context(), params.WorkspaceID, opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "updated vcs settings")
	http.Redirect(w, r, path.Resource(resource.Action("edit-vcs"), ws.ID), http.StatusFound)
}

func (h *Handlers) editAdvanced(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.RenderPage(
		editAdvanced(workspaceID, h.Authorizer),
		"edit advanced | "+workspaceID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithSideMenu(helpers.WorkspaceSettingsMenu(ws.ID)),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Advanced"},
		),
	)
}

func (h *Handlers) createTag(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID *resource.TfeID `schema:"workspace_id,required"`
		TagName     *string         `schema:"tag_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.Client.AddTags(r.Context(), *params.WorkspaceID, []workspace.TagSpec{{Name: *params.TagName}})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "created tag: "+*params.TagName)
	http.Redirect(w, r, path.Get(params.WorkspaceID), http.StatusFound)
}

func (h *Handlers) deleteTag(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID *resource.TfeID `schema:"workspace_id,required"`
		TagName     *string         `schema:"tag_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.Client.RemoveTags(r.Context(), *params.WorkspaceID, []workspace.TagSpec{{Name: *params.TagName}})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "removed tag: "+*params.TagName)
	http.Redirect(w, r, path.Get(params.WorkspaceID), http.StatusFound)
}

// filterUnassigned removes from the list of teams those that are part of the
// policy, i.e. those that have been assigned a permission.
//
// NOTE: the owners team is always removed because by default it is assigned the
// admin role.
func filterUnassigned(policy workspace.Policy, teams []*team.Team) (unassigned []*team.Team) {
	assigned := make(map[resource.TfeID]struct{}, len(teams))
	for _, p := range policy.Permissions {
		assigned[p.TeamID] = struct{}{}
	}
	for _, t := range teams {
		if t.Name == "owners" {
			continue
		}
		if _, ok := assigned[t.ID]; !ok {
			unassigned = append(unassigned, t)
		}
	}
	return
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
