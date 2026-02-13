package ui

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
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

// addWorkspaceHandlers registers workspace UI handlers with the router
func addWorkspaceHandlers(r *mux.Router, h *Handlers) {
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

	r.HandleFunc("/workspaces/{workspace_id}/create-tag", h.createTag).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/delete-tag", h.deleteTag).Methods("POST")

}

func (h *Handlers) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var params struct {
		workspace.ListOptions
		StatusFilterVisible bool `schema:"status_filter_visible"`
		TagFilterVisible    bool `schema:"tag_filter_visible"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// retrieve all tags for listing in tag filter
	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Tag], error) {
		return h.Workspaces.ListTags(r.Context(), *params.Organization, workspace.ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	tagStrings := make([]string, len(tags))
	for i, tag := range tags {
		tagStrings[i] = tag.Name
	}

	page, err := h.Workspaces.List(r.Context(), params.ListOptions)
	if err != nil {
		html.Error(r, w, err.Error())
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

	h.renderPage(
		h.templates.workspaceList(props),
		"workspaces",
		w,
		r,
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Workspaces"},
		),
		withOrganization(*params.Organization),
		withContentActions(workspaceListActions(
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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	h.renderPage(
		h.templates.workspaceNew(params.Organization),
		"new workspace",
		w,
		r,
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "workspaces", Link: paths.Workspaces(params.Organization)},
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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Create(r.Context(), workspace.CreateOptions{
		Name:         params.Name,
		Organization: params.Organization,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "created workspace: "+ws.Name)
	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *Handlers) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Get(r.Context(), id)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	user, err := user.UserFromContext(r.Context())
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	var provider *vcs.Provider
	if ws.Connection != nil {
		provider, err = h.VCSProviders.Get(r.Context(), ws.Connection.VCSProviderID)
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
	}

	// retrieve tags that are available to be assigned to the workspace
	// (excluding those already assigned to the workspace).
	var availableTags []string
	{
		tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Tag], error) {
			return h.Workspaces.ListTags(r.Context(), ws.Organization, workspace.ListTagsOptions{
				PageOptions: opts,
			})
		})
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
		names := internal.Map(tags, func(t *workspace.Tag) string {
			return t.Name
		})
		availableTags = internal.Diff(names, ws.Tags)
	}

	lockInfo, err := h.lockButtonHelper(r.Context(), ws, user)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	// Generate component for latest run if workspace has one.
	var latestRunTable templ.Component
	if ws.LatestRun != nil {
		run, err := h.Runs.Get(r.Context(), ws.LatestRun.ID)
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
		latestRunTable = h.templates.singleRunTable(run)
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
			Action:      templ.SafeURL(paths.CreateTagWorkspace(ws.ID)),
			Placeholder: "Add tags",
			Width:       helpers.NarrowDropDown,
		},
		latestRunTable: latestRunTable,
	}
	h.renderPage(
		h.templates.workspaceGet(props),
		"workspaces",
		w,
		r,
		withWorkspace(ws),
		withBreadcrumbs(
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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.GetByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *Handlers) editWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Get(r.Context(), workspaceID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	policy, err := h.Workspaces.GetWorkspacePolicy(r.Context(), workspaceID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	// Get teams for populating team permissions
	teams, err := h.Teams.List(r.Context(), ws.Organization)
	if err != nil {
		html.Error(r, w, err.Error())
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

	var provider *vcs.Provider
	if ws.Connection != nil {
		provider, err = h.VCSProviders.Get(r.Context(), ws.Connection.VCSProviderID)
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
	}

	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Tag], error) {
		return h.Workspaces.ListTags(r.Context(), ws.Organization, workspace.ListTagsOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	poolsURL := paths.PoolsWorkspace(workspaceID)
	if ws.AgentPoolID != nil {
		poolsURL += "?agent_pool_id=" + ws.AgentPoolID.String()
	}

	// Construct list of engines for template
	engineSelectorProps := engineSelectorProps{
		engines: make([]engineSelectorEngine, 2),
	}
	for i, engine := range engine.Engines() {
		engineSelectorProps.engines[i].name = engine.String()
		if engine.String() == ws.Engine.String() {
			engineSelectorProps.current = engine.String()
			engineSelectorProps.engines[i].latest = ws.EngineVersion.Latest
		}
		// Offer the user the latest available version for an engine if:
		// (a): it's not the current engine
		// (b): it's currently set to track the latest version.
		if engineSelectorProps.current != engine.String() || engineSelectorProps.engines[i].latest {
			latest, _, err := h.EngineService.GetLatest(r.Context(), engine)
			if err != nil {
				html.Error(r, w, err.Error())
				return
			}
			engineSelectorProps.engines[i].version = latest
		} else {
			engineSelectorProps.engines[i].version = ws.EngineVersion.String()
		}
	}

	props := workspaceEditProps{
		ws:         ws,
		assigned:   perms,
		unassigned: filterUnassigned(policy, teams),
		engines:    engineSelectorProps,
		roles: []authz.Role{
			authz.WorkspaceReadRole,
			authz.WorkspacePlanRole,
			authz.WorkspaceWriteRole,
			authz.WorkspaceAdminRole,
		},
		vcsProvider:        provider,
		unassignedTags:     internal.Diff(tagNames, ws.Tags),
		vcsTagRegexDefault: vcsTagRegexDefault,
		vcsTagRegexPrefix:  vcsTagRegexPrefix,
		vcsTagRegexSuffix:  vcsTagRegexSuffix,
		vcsTagRegexCustom:  vcsTagRegexCustom,
		vcsTriggerAlways:   VCSTriggerAlways,
		vcsTriggerPatterns: VCSTriggerPatterns,
		vcsTriggerTags:     VCSTriggerTags,
		canUpdateWorkspace: h.Authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, ws.ID),
		canDeleteWorkspace: h.Authorizer.CanAccess(r.Context(), authz.DeleteWorkspaceAction, ws.ID),
		poolsURL:           poolsURL,
	}
	h.renderPage(
		h.templates.workspaceEdit(props),
		"edit | "+ws.ID.String(),
		w,
		r,
		withWorkspace(ws),
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Settings"},
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
		Engine                *engine.Engine          `schema:"engine"`
		LatestEngineVersion   bool                    `schema:"latest_engine_version"`
		SpecificEngineVersion *workspace.Version      `schema:"specific_engine_version"`
		WorkingDirectory      string                  `schema:"working_directory"`
		WorkspaceID           resource.TfeID          `schema:"workspace_id,required"`
		GlobalRemoteState     bool                    `schema:"global_remote_state"`

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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// get workspace before updating to determine if it is connected or not.
	ws, err := h.Workspaces.Get(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	opts := workspace.UpdateOptions{
		AutoApply:          &params.AutoApply,
		Name:               &params.Name,
		Description:        &params.Description,
		ExecutionMode:      &params.ExecutionMode,
		Engine:             params.Engine,
		WorkingDirectory:   &params.WorkingDirectory,
		GlobalRemoteState:  &params.GlobalRemoteState,
		SpeculativeEnabled: &params.SpeculativeEnabled,
	}
	if params.LatestEngineVersion {
		opts.EngineVersion = &workspace.Version{Latest: true}
	} else {
		opts.EngineVersion = params.SpecificEngineVersion
	}
	if ws.Connection != nil {
		// workspace is connected, so set connection fields
		opts.ConnectOptions = &workspace.ConnectOptions{
			AllowCLIApply: &params.AllowCLIApply,
			Branch:        &params.VCSBranch,
		}
		switch params.VCSTriggerStrategy {
		case VCSTriggerAlways:
			opts.AlwaysTrigger = new(true)
		case VCSTriggerPatterns:
			err := json.Unmarshal([]byte(params.TriggerPatternsJSON), &opts.TriggerPatterns)
			if err != nil {
				html.Error(r, w, err.Error())
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
	if params.ExecutionMode == workspace.AgentExecutionMode {
		opts.AgentPoolID = params.AgentPoolID
	}

	ws, err = h.Workspaces.Update(r.Context(), params.WorkspaceID, opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, paths.EditWorkspace(ws.ID), http.StatusFound)
}

func (h *Handlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Delete(r.Context(), workspaceID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "deleted workspace: "+ws.Name)
	http.Redirect(w, r, paths.Workspaces(ws.Organization), http.StatusFound)
}

func (h *Handlers) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Lock(r.Context(), id, nil)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *Handlers) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Unlock(r.Context(), workspaceID, nil, false)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *Handlers) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Unlock(r.Context(), workspaceID, nil, true)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID), http.StatusFound)
}

func (h *Handlers) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Get(r.Context(), workspaceID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	providers, err := h.VCSProviders.List(r.Context(), ws.Organization)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	h.renderPage(
		h.templates.listVCSProviders(ws, providers),
		"list vcs providers | "+ws.ID.String(),
		w,
		r,
		withWorkspace(ws),
		withBreadcrumbs(
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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.Get(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	client, err := h.VCSProviders.Get(r.Context(), params.VCSProviderID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	repos, err := client.ListRepositories(r.Context(), vcs.ListRepositoriesOptions{
		PageSize: resource.MaxPageSize,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	h.renderPage(
		h.templates.listVCSRepos(ws, params.VCSProviderID, repos),
		"list vcs repos | "+ws.ID.String(),
		w,
		r,
		withWorkspace(ws),
		withBreadcrumbs(
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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	_, err := h.Workspaces.Update(r.Context(), params.WorkspaceID, workspace.UpdateOptions{
		ConnectOptions: &workspace.ConnectOptions{
			VCSProviderID: params.VCSProviderID,
			RepoPath:      params.RepoPath,
		},
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}

func (h *Handlers) disconnect(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	_, err = h.Workspaces.Update(r.Context(), workspaceID, workspace.UpdateOptions{
		Disconnect: true,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, paths.Workspace(workspaceID), http.StatusFound)
}

func (h *Handlers) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		TeamID      resource.TfeID `schema:"team_id,required"`
		Role        string         `schema:"role,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	role, err := authz.WorkspaceRoleFromString(params.Role)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	err = h.Workspaces.SetPermission(r.Context(), params.WorkspaceID, params.TeamID, role)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "updated workspace permissions")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID), http.StatusFound)
}

func (h *Handlers) unsetWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		TeamID      resource.TfeID `schema:"team_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.Workspaces.UnsetPermission(r.Context(), params.WorkspaceID, params.TeamID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "deleted workspace permission")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID), http.StatusFound)
}

func (h *Handlers) createTag(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID *resource.TfeID `schema:"workspace_id,required"`
		TagName     *string         `schema:"tag_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.Workspaces.AddTags(r.Context(), *params.WorkspaceID, []workspace.TagSpec{{Name: *params.TagName}})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "created tag: "+*params.TagName)
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}

func (h *Handlers) deleteTag(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID *resource.TfeID `schema:"workspace_id,required"`
		TagName     *string         `schema:"tag_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.Workspaces.RemoveTags(r.Context(), *params.WorkspaceID, []workspace.TagSpec{{Name: *params.TagName}})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "removed tag: "+*params.TagName)
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
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
