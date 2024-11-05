package workspace

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
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
		html.Renderer

		teams        webTeamClient
		vcsproviders webVCSProvidersClient
		client       webClient
	}

	webTeamClient interface {
		List(context.Context, string) ([]*team.Team, error)
	}

	webVCSProvidersClient interface {
		Get(ctx context.Context, providerID resource.ID) (*vcsprovider.VCSProvider, error)
		List(context.Context, string) ([]*vcsprovider.VCSProvider, error)

		GetVCSClient(ctx context.Context, providerID resource.ID) (vcs.Client, error)
	}

	// webClient provides web handlers with access to the workspace service
	webClient interface {
		Create(ctx context.Context, opts CreateOptions) (*Workspace, error)
		Get(ctx context.Context, workspaceID resource.ID) (*Workspace, error)
		GetByName(ctx context.Context, organization, workspace string) (*Workspace, error)
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error)
		Update(ctx context.Context, workspaceID resource.ID, opts UpdateOptions) (*Workspace, error)
		Delete(ctx context.Context, workspaceID resource.ID) (*Workspace, error)
		Lock(ctx context.Context, workspaceID resource.ID, runID *string) (*Workspace, error)
		Unlock(ctx context.Context, workspaceID resource.ID, runID *string, force bool) (*Workspace, error)

		AddTags(ctx context.Context, workspaceID resource.ID, tags []TagSpec) error
		RemoveTags(ctx context.Context, workspaceID resource.ID, tags []TagSpec) error
		ListTags(ctx context.Context, organization string, opts ListTagsOptions) (*resource.Page[*Tag], error)

		GetPolicy(ctx context.Context, workspaceID resource.ID) (authz.WorkspacePolicy, error)
		SetPermission(ctx context.Context, workspaceID, teamID resource.ID, role rbac.Role) error
		UnsetPermission(ctx context.Context, workspaceID, teamID resource.ID) error
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

	workspaces, err := h.client.List(r.Context(), ListOptions{
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
		return h.client.ListTags(r.Context(), *params.Organization, ListTagsOptions{
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

	user, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		organization.OrganizationPage
		*resource.Page[*Workspace]
		TagFilters         map[string]bool
		Search             string
		CanCreateWorkspace bool
	}{
		OrganizationPage:   organization.NewPage(r, "workspaces", *params.Organization),
		CanCreateWorkspace: user.CanAccessOrganization(rbac.CreateTeamAction, *params.Organization),
		Page:               workspaces,
		TagFilters:         tagfilters(),
		Search:             params.Search,
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

	ws, err := h.client.Create(r.Context(), CreateOptions{
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
	http.Redirect(w, r, paths.Workspace(ws.ID.String()), http.StatusFound)
}

func (h *webHandlers) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Get(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	policy, err := h.client.GetPolicy(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var provider *vcsprovider.VCSProvider
	if ws.Connection != nil {
		provider, err = h.vcsproviders.Get(r.Context(), ws.Connection.VCSProviderID)
		if err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Tag], error) {
		return h.client.ListTags(r.Context(), ws.Organization, ListTagsOptions{
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
		VCSProvider        *vcsprovider.VCSProvider
		CanApply           bool
		CanAddTags         bool
		CanRemoveTags      bool
		CanCreateRun       bool
		CanLockWorkspace   bool
		CanUnlockWorkspace bool
		CanUpdateWorkspace bool
		UnassignedTags     []string
		TagsDropdown       html.DropdownUI
	}{
		WorkspacePage:      NewPage(r, ws.Name, ws),
		LockButton:         lockButtonHelper(ws, policy, user),
		VCSProvider:        provider,
		CanApply:           user.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		CanAddTags:         user.CanAccessWorkspace(rbac.AddTagsAction, policy),
		CanRemoveTags:      user.CanAccessWorkspace(rbac.RemoveTagsAction, policy),
		CanCreateRun:       user.CanAccessWorkspace(rbac.CreateRunAction, policy),
		CanLockWorkspace:   user.CanAccessWorkspace(rbac.LockWorkspaceAction, policy),
		CanUnlockWorkspace: user.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy),
		CanUpdateWorkspace: user.CanAccessWorkspace(rbac.UpdateWorkspaceAction, policy),
		TagsDropdown: html.DropdownUI{
			Name:        "tag_name",
			Available:   internal.Diff(getTagNames(), ws.Tags),
			Existing:    ws.Tags,
			Action:      paths.CreateTagWorkspace(ws.ID.String()),
			Placeholder: "Add tags",
			Width:       html.NarrowDropDown,
		},
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

	ws, err := h.client.GetByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID.String()), http.StatusFound)
}

func (h *webHandlers) editWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspace, err := h.client.Get(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	policy, err := h.client.GetPolicy(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get teams for populating team permissions
	teams, err := h.teams.List(r.Context(), workspace.Organization)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// want current policy permissions to include not only team ID but team name
	// too for user's benefit
	type perm struct {
		Role rbac.Role
		Team *team.Team
	}
	perms := make([]perm, len(policy.Permissions))
	for i, pp := range policy.Permissions {
		// get team name corresponding to team ID
		for _, t := range teams {
			if t.ID == pp.TeamID {
				perms[i] = perm{
					Role: pp.Role,
					Team: t,
				}
				break
			}
		}
	}

	var provider *vcsprovider.VCSProvider
	if workspace.Connection != nil {
		provider, err = h.vcsproviders.Get(r.Context(), workspace.Connection.VCSProviderID)
		if err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tags, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Tag], error) {
		return h.client.ListTags(r.Context(), workspace.Organization, ListTagsOptions{
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
		Assigned           []perm
		Unassigned         []*team.Team
		Roles              []rbac.Role
		VCSProvider        *vcsprovider.VCSProvider
		UnassignedTags     []string
		CanUpdateWorkspace bool
		CanDeleteWorkspace bool
		VCSTagRegexDefault string
		VCSTagRegexPrefix  string
		VCSTagRegexSuffix  string
		VCSTagRegexCustom  string
		VCSTriggerAlways   string
		VCSTriggerPatterns string
		VCSTriggerTags     string
	}{
		WorkspacePage: NewPage(r, "edit | "+workspace.ID.String(), workspace),
		Assigned:      perms,
		Unassigned:    filterUnassigned(policy, teams),
		Roles: []rbac.Role{
			rbac.WorkspaceReadRole,
			rbac.WorkspacePlanRole,
			rbac.WorkspaceWriteRole,
			rbac.WorkspaceAdminRole,
		},
		VCSProvider:        provider,
		UnassignedTags:     internal.Diff(getTagNames(), workspace.Tags),
		VCSTagRegexDefault: vcsTagRegexDefault,
		VCSTagRegexPrefix:  vcsTagRegexPrefix,
		VCSTagRegexSuffix:  vcsTagRegexSuffix,
		VCSTagRegexCustom:  vcsTagRegexCustom,
		VCSTriggerAlways:   VCSTriggerAlways,
		VCSTriggerPatterns: VCSTriggerPatterns,
		VCSTriggerTags:     VCSTriggerTags,
		CanUpdateWorkspace: user.CanAccessWorkspace(rbac.UpdateWorkspaceAction, policy),
		CanDeleteWorkspace: user.CanAccessWorkspace(rbac.DeleteWorkspaceAction, policy),
	})
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
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// get workspace before updating to determine if it is connected or not.
	ws, err := h.client.Get(r.Context(), params.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := UpdateOptions{
		AutoApply:         &params.AutoApply,
		Name:              &params.Name,
		Description:       &params.Description,
		ExecutionMode:     &params.ExecutionMode,
		TerraformVersion:  &params.TerraformVersion,
		WorkingDirectory:  &params.WorkingDirectory,
		GlobalRemoteState: &params.GlobalRemoteState,
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
				h.Error(w, err.Error(), http.StatusInternalServerError)
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
		opts.AgentPoolID = &params.AgentPoolID
	}

	ws, err = h.client.Update(r.Context(), params.WorkspaceID, opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated workspace")
	// User may have updated workspace name so path references updated workspace
	http.Redirect(w, r, paths.EditWorkspace(ws.ID.String()), http.StatusFound)
}

func (h *webHandlers) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Delete(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace: "+ws.Name)
	http.Redirect(w, r, paths.Workspaces(ws.Organization), http.StatusFound)
}

func (h *webHandlers) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Lock(r.Context(), id, nil)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(ws.ID.String()), http.StatusFound)
}

func (h *webHandlers) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Unlock(r.Context(), workspaceID, nil, false)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID.String()), http.StatusFound)
}

func (h *webHandlers) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Unlock(r.Context(), workspaceID, nil, true)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Workspace(ws.ID.String()), http.StatusFound)
}

func (h *webHandlers) listWorkspaceVCSProviders(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Get(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	providers, err := h.vcsproviders.List(r.Context(), ws.Organization)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("workspace_vcs_provider_list.tmpl", w, struct {
		WorkspacePage
		Items []*vcsprovider.VCSProvider
	}{
		WorkspacePage: NewPage(r, "list vcs providers | "+ws.ID.String(), ws),
		Items:         providers,
	})
}

func (h *webHandlers) listWorkspaceVCSRepos(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID   resource.ID `schema:"workspace_id,required"`
		VCSProviderID resource.ID `schema:"vcs_provider_id,required"`

		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.client.Get(r.Context(), params.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client, err := h.vcsproviders.GetVCSClient(r.Context(), params.VCSProviderID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repos, err := client.ListRepositories(r.Context(), vcs.ListRepositoriesOptions{
		PageSize: html.PageSize,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("workspace_vcs_repo_list.tmpl", w, struct {
		WorkspacePage
		Repos         []string
		VCSProviderID resource.ID
	}{
		WorkspacePage: NewPage(r, "list vcs repos | "+ws.ID.String(), ws),
		Repos:         repos,
		VCSProviderID: params.VCSProviderID,
	})
}

func (h *webHandlers) connect(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID   resource.ID  `schema:"workspace_id,required"`
		RepoPath      *string      `schema:"identifier,required"`
		VCSProviderID *resource.ID `schema:"vcs_provider_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err := h.client.Update(r.Context(), params.WorkspaceID, UpdateOptions{
		ConnectOptions: &ConnectOptions{
			VCSProviderID: params.VCSProviderID,
			RepoPath:      params.RepoPath,
		},
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "connected workspace to repo")
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID.String()), http.StatusFound)
}

func (h *webHandlers) disconnect(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err = h.client.Update(r.Context(), workspaceID, UpdateOptions{
		Disconnect: true,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "disconnected workspace from repo")
	http.Redirect(w, r, paths.Workspace(workspaceID.String()), http.StatusFound)
}

func (h *webHandlers) setWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.ID `schema:"workspace_id,required"`
		TeamID      resource.ID `schema:"team_id,required"`
		Role        string      `schema:"role,required"`
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

	err = h.client.SetPermission(r.Context(), params.WorkspaceID, params.TeamID, role)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "updated workspace permissions")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID.String()), http.StatusFound)
}

func (h *webHandlers) unsetWorkspacePermission(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.ID `schema:"workspace_id,required"`
		TeamID      resource.ID `schema:"team_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.client.UnsetPermission(r.Context(), params.WorkspaceID, params.TeamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted workspace permission")
	http.Redirect(w, r, paths.EditWorkspace(params.WorkspaceID.String()), http.StatusFound)
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
