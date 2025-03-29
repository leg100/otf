package workspace

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/components"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

const (
	// give user choice of pre-defined regexes for matching vcs tags
	vcsTagRegexDefault = `^\d+\.\d+\.\d+$`
	vcsTagRegexPrefix  = `\d+\.\d+\.\d+$`
	vcsTagRegexSuffix  = `^\d+\.\d+\.\d+`
	// this is a 'magic string' that indicates a custom regex has been
	// supplied in another variable
	vcsTagRegexCustom = `custom`

	//
	// VCS trigger strategies to present to the user.
	//
	// every vcs event trigger runs
	VCSTriggerAlways string = "always"
	// only vcs events with changed files matching a set of glob patterns
	// triggers run
	VCSTriggerPatterns string = "patterns"
	// only push tag vcs events trigger runs
	VCSTriggerTags string = "tags"
)

type (
	webHandlers struct {
		*uiHelpers

		teams                webTeamClient
		vcsproviders         webVCSProvidersClient
		client               webClient
		authorizer           webAuthorizer
		websocketListHandler *components.WebsocketListHandler[*Workspace, ListOptions]
	}

	webTeamClient interface {
		List(context.Context, resource.OrganizationName) ([]*team.Team, error)
	}

	webVCSProvidersClient interface {
		Get(ctx context.Context, providerID resource.ID) (*vcsprovider.VCSProvider, error)
		List(context.Context, resource.OrganizationName) ([]*vcsprovider.VCSProvider, error)

		GetVCSClient(ctx context.Context, providerID resource.ID) (vcs.Client, error)
	}

	webAuthorizer interface {
		CanAccess(context.Context, authz.Action, *authz.AccessRequest) bool
	}

	// webClient provides web handlers with access to the workspace service
	webClient interface {
		Create(ctx context.Context, opts CreateOptions) (*Workspace, error)
		Get(ctx context.Context, workspaceID resource.ID) (*Workspace, error)
		GetByName(ctx context.Context, organization resource.OrganizationName, workspace string) (*Workspace, error)
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error)
		Update(ctx context.Context, workspaceID resource.ID, opts UpdateOptions) (*Workspace, error)
		Delete(ctx context.Context, workspaceID resource.ID) (*Workspace, error)
		Lock(ctx context.Context, workspaceID resource.ID, runID resource.ID) (*Workspace, error)
		Unlock(ctx context.Context, workspaceID resource.ID, runID resource.ID, force bool) (*Workspace, error)

		AddTags(ctx context.Context, workspaceID resource.ID, tags []TagSpec) error
		RemoveTags(ctx context.Context, workspaceID resource.ID, tags []TagSpec) error
		ListTags(ctx context.Context, organization resource.OrganizationName, opts ListTagsOptions) (*resource.Page[*Tag], error)

		GetWorkspacePolicy(ctx context.Context, workspaceID resource.ID) (authz.WorkspacePolicy, error)
		SetPermission(ctx context.Context, workspaceID, teamID resource.ID, role authz.Role) error
		UnsetPermission(ctx context.Context, workspaceID, teamID resource.ID) error
	}
)

func newWebHandlers(service *Service, opts Options) *webHandlers {
	return &webHandlers{
		authorizer:   opts.Authorizer,
		teams:        opts.TeamService,
		vcsproviders: opts.VCSProviderService,
		client:       service,
		uiHelpers: &uiHelpers{
			service:    opts.UserService,
			authorizer: opts.Authorizer,
		},
		websocketListHandler: &components.WebsocketListHandler[*Workspace, ListOptions]{
			Logger:    opts.Logger,
			Client:    service,
			Populator: &table{},
		},
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
	if websocket.IsWebSocketUpgrade(r) {
		h.websocketListHandler.Handler(w, r)
		return
	}

	var params struct {
		ListOptions
		StatusFilterVisible bool `schema:"status_filter_visible"`
		TagFilterVisible    bool `schema:"tag_filter_visible"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// retrieve all tags and create map, with each entry determining whether
	// listing is currently filtered by the tag or not.
	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Tag], error) {
		return h.client.ListTags(r.Context(), *params.Organization, ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tagMap := make(map[string]bool, len(tags))
	for _, t := range tags {
		tagMap[t.Name] = false
		for _, f := range params.Tags {
			if t.Name == f {
				tagMap[t.Name] = true
				break
			}
		}
	}

	props := listProps{
		organization:        *params.Organization,
		tags:                tagMap,
		search:              params.Search,
		status:              params.Status,
		statusFilterVisible: params.StatusFilterVisible,
		tagFilterVisible:    params.TagFilterVisible,
		canCreate: h.authorizer.CanAccess(
			r.Context(),
			authz.CreateWorkspaceAction,
			&authz.AccessRequest{Organization: params.Organization},
		),
		pageOptions: params.PageOptions,
	}

	html.Render(list(props), w, r)
}

func (h *webHandlers) newWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization resource.OrganizationName `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	html.Render(new(params.Organization), w, r)
}

func (h *webHandlers) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         *string                    `schema:"name,required"`
		Organization *resource.OrganizationName `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Create(r.Context(), CreateOptions{
		Name:         params.Name,
		Organization: params.Organization,
	})
	if err == internal.ErrResourceAlreadyExists {
		html.FlashError(w, "workspace already exists: "+*params.Name)
		http.Redirect(w, r, paths.NewWorkspace(params.Organization), http.StatusFound)
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
	id, err := decode.TfeID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Get(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := user.UserFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var provider *vcsprovider.VCSProvider
	if ws.Connection != nil {
		provider, err = h.vcsproviders.Get(r.Context(), ws.Connection.VCSProviderID)
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Tag], error) {
		return h.client.ListTags(r.Context(), ws.Organization, ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	getTagNames := func() (names []string) {
		for _, t := range tags {
			names = append(names, t.Name)
		}
		return
	}

	lockButton, err := h.lockButtonHelper(r.Context(), ws, user)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := getProps{
		ws:                 ws,
		button:             lockButton,
		vcsProvider:        provider,
		canApply:           h.authorizer.CanAccess(r.Context(), authz.ApplyRunAction, &authz.AccessRequest{ID: ws.ID}),
		canAddTags:         h.authorizer.CanAccess(r.Context(), authz.AddTagsAction, &authz.AccessRequest{ID: ws.ID}),
		canRemoveTags:      h.authorizer.CanAccess(r.Context(), authz.RemoveTagsAction, &authz.AccessRequest{ID: ws.ID}),
		canCreateRun:       h.authorizer.CanAccess(r.Context(), authz.CreateRunAction, &authz.AccessRequest{ID: ws.ID}),
		canLockWorkspace:   h.authorizer.CanAccess(r.Context(), authz.LockWorkspaceAction, &authz.AccessRequest{ID: ws.ID}),
		canUnlockWorkspace: h.authorizer.CanAccess(r.Context(), authz.UnlockWorkspaceAction, &authz.AccessRequest{ID: ws.ID}),
		canUpdateWorkspace: h.authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, &authz.AccessRequest{ID: ws.ID}),
		tagsDropdown: components.SearchDropdownProps{
			Name:        "tag_name",
			Available:   internal.Diff(getTagNames(), ws.Tags),
			Existing:    ws.Tags,
			Action:      templ.SafeURL(paths.CreateTagWorkspace(ws.ID)),
			Placeholder: "Add tags",
			Width:       components.NarrowDropDown,
		},
	}
	html.Render(get(props), w, r)
}

func (h *webHandlers) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string                    `schema:"workspace_name,required"`
		Organization resource.OrganizationName `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.GetByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) editWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.TfeID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspace, err := h.client.Get(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	policy, err := h.client.GetWorkspacePolicy(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get teams for populating team permissions
	teams, err := h.teams.List(r.Context(), workspace.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// want current policy permissions to include not only team ID but team name
	// too for user's benefit
	perms := make([]perm, len(policy.Permissions))
	for i, pp := range policy.Permissions {
		// get team name corresponding to team ID
		for _, t := range teams {
			if t.ID == pp.TeamID {
				perms[i] = perm{
					role: pp.Role,
					team: t,
				}
				break
			}
		}
	}

	var provider *vcsprovider.VCSProvider
	if workspace.Connection != nil {
		provider, err = h.vcsproviders.Get(r.Context(), workspace.Connection.VCSProviderID)
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Tag], error) {
		return h.client.ListTags(r.Context(), workspace.Organization, ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	getTagNames := func() (names []string) {
		for _, t := range tags {
			names = append(names, t.Name)
		}
		return
	}

	poolsURL := paths.PoolsWorkspace(workspaceID)
	if workspace.AgentPoolID != nil {
		poolsURL += "?agent_pool_id=" + workspace.AgentPoolID.String()
	}

	props := editProps{
		ws:         workspace,
		assigned:   perms,
		unassigned: filterUnassigned(policy, teams),
		roles: []authz.Role{
			authz.WorkspaceReadRole,
			authz.WorkspacePlanRole,
			authz.WorkspaceWriteRole,
			authz.WorkspaceAdminRole,
		},
		vcsProvider:        provider,
		unassignedTags:     internal.Diff(getTagNames(), workspace.Tags),
		vcsTagRegexDefault: vcsTagRegexDefault,
		vcsTagRegexPrefix:  vcsTagRegexPrefix,
		vcsTagRegexSuffix:  vcsTagRegexSuffix,
		vcsTagRegexCustom:  vcsTagRegexCustom,
		vcsTriggerAlways:   VCSTriggerAlways,
		vcsTriggerPatterns: VCSTriggerPatterns,
		vcsTriggerTags:     VCSTriggerTags,
		canUpdateWorkspace: h.authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, &authz.AccessRequest{ID: workspace.ID}),
		canDeleteWorkspace: h.authorizer.CanAccess(r.Context(), authz.DeleteWorkspaceAction, &authz.AccessRequest{ID: workspace.ID}),
		poolsURL:           poolsURL,
	}
	html.Render(edit(props), w, r)
}

func (h *webHandlers) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		AgentPoolID       resource.ID `schema:"agent_pool_id"`
		AutoApply         bool        `schema:"auto_apply"`
		Name              string
		Description       string
		ExecutionMode     ExecutionMode `schema:"execution_mode"`
		TerraformVersion  string        `schema:"terraform_version"`
		WorkingDirectory  string        `schema:"working_directory"`
		WorkspaceID       resource.ID   `schema:"workspace_id,required"`
		GlobalRemoteState bool          `schema:"global_remote_state"`

		// VCS connection
		VCSTriggerStrategy  string `schema:"vcs_trigger"`
		TriggerPatternsJSON string `schema:"trigger_patterns"`
		VCSBranch           string `schema:"vcs_branch"`
		PredefinedTagsRegex string `schema:"tags_regex"`
		CustomTagsRegex     string `schema:"custom_tags_regex"`
		AllowCLIApply       bool   `schema:"allow_cli_apply"`
		SpeculativeEnabled  bool   `schema:"speculative_enabled"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// get workspace before updating to determine if it is connected or not.
	ws, err := h.client.Get(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := UpdateOptions{
		AutoApply:          &params.AutoApply,
		Name:               &params.Name,
		Description:        &params.Description,
		ExecutionMode:      &params.ExecutionMode,
		TerraformVersion:   &params.TerraformVersion,
		WorkingDirectory:   &params.WorkingDirectory,
		GlobalRemoteState:  &params.GlobalRemoteState,
		SpeculativeEnabled: &params.SpeculativeEnabled,
	}
	if ws.Connection != nil {
		// workspace is connected, so set connection fields
		opts.ConnectOptions = &ConnectOptions{
			AllowCLIApply: &params.AllowCLIApply,
			Branch:        &params.VCSBranch,
		}
		switch params.VCSTriggerStrategy {
		case VCSTriggerAlways:
			opts.AlwaysTrigger = internal.Bool(true)
		case VCSTriggerPatterns:
			err := json.Unmarshal([]byte(params.TriggerPatternsJSON), &opts.TriggerPatterns)
			if err != nil {
				html.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case VCSTriggerTags:
			if params.PredefinedTagsRegex == vcsTagRegexCustom {
				opts.ConnectOptions.TagsRegex = &params.CustomTagsRegex
			} else {
				opts.ConnectOptions.TagsRegex = &params.PredefinedTagsRegex
			}
		}
	}
	// only set agent pool ID if execution mode is set to agent
	if params.ExecutionMode == AgentExecutionMode {
		opts.AgentPoolID = params.AgentPoolID
	}

	ws, err = h.client.Update(r.Context(), params.WorkspaceID, opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, paths.EditWorkspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.TfeID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Delete(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace: "+ws.Name)
	http.Redirect(w, r, paths.Workspaces(ws.Organization), http.StatusFound)
}

func (h *webHandlers) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.TfeID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Lock(r.Context(), id, nil)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.TfeID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Unlock(r.Context(), workspaceID, nil, false)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.TfeID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Unlock(r.Context(), workspaceID, nil, true)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *webHandlers) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.TfeID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Get(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	providers, err := h.vcsproviders.List(r.Context(), ws.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.Render(listVCSProviders(ws, providers), w, r)
}

func (h *webHandlers) listWorkspaceVCSRepos(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID   resource.TfeID `schema:"workspace_id,required"`
		VCSProviderID resource.TfeID `schema:"vcs_provider_id,required"`

		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Get(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client, err := h.vcsproviders.GetVCSClient(r.Context(), params.VCSProviderID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repos, err := client.ListRepositories(r.Context(), vcs.ListRepositoriesOptions{
		PageSize: resource.MaxPageSize,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.Render(listVCSRepos(ws, params.VCSProviderID, repos), w, r)
}

func (h *webHandlers) connect(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID   resource.TfeID `schema:"workspace_id,required"`
		RepoPath      *string        `schema:"identifier,required"`
		VCSProviderID resource.TfeID `schema:"vcs_provider_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err := h.client.Update(r.Context(), params.WorkspaceID, UpdateOptions{
		ConnectOptions: &ConnectOptions{
			VCSProviderID: params.VCSProviderID,
			RepoPath:      params.RepoPath,
		},
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) disconnect(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.TfeID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err = h.client.Update(r.Context(), workspaceID, UpdateOptions{
		Disconnect: true,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, paths.Workspace(workspaceID), http.StatusFound)
}

func (h *webHandlers) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		TeamID      resource.TfeID `schema:"team_id,required"`
		Role        string         `schema:"role,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	role, err := authz.WorkspaceRoleFromString(params.Role)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.client.SetPermission(r.Context(), params.WorkspaceID, params.TeamID, role)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "updated workspace permissions")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) unsetWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		TeamID      resource.TfeID `schema:"team_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.client.UnsetPermission(r.Context(), params.WorkspaceID, params.TeamID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
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
func filterUnassigned(policy authz.WorkspacePolicy, teams []*team.Team) (unassigned []*team.Team) {
	assigned := make(map[resource.ID]struct{}, len(teams))
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
