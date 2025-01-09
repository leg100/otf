package runner

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
	workspacepkg "github.com/leg100/otf/internal/workspace"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	html.Renderer

	svc        webClient
	workspaces *workspacepkg.Service
	logger     logr.Logger
	authorizer webAuthorizer
}

// webClient gives web handlers access to the agents service endpoints
type webClient interface {
	CreateAgentPool(ctx context.Context, opts CreateAgentPoolOptions) (*Pool, error)
	GetAgentPool(ctx context.Context, poolID resource.ID) (*Pool, error)
	updateAgentPool(ctx context.Context, poolID resource.ID, opts updatePoolOptions) (*Pool, error)
	listAgentPoolsByOrganization(ctx context.Context, organization string, opts listPoolOptions) ([]*Pool, error)
	deleteAgentPool(ctx context.Context, poolID resource.ID) (*Pool, error)

	register(ctx context.Context, opts registerOptions) (*RunnerMeta, error)
	listRunners(ctx context.Context) ([]*RunnerMeta, error)
	listRunnersByOrganization(ctx context.Context, organization string) ([]*RunnerMeta, error)
	listRunnersByPool(ctx context.Context, poolID resource.ID) ([]*RunnerMeta, error)
	listServerRunners(ctx context.Context) ([]*RunnerMeta, error)

	CreateAgentToken(ctx context.Context, poolID resource.ID, opts CreateAgentTokenOptions) (*agentToken, []byte, error)
	GetAgentToken(ctx context.Context, tokenID resource.ID) (*agentToken, error)
	ListAgentTokens(ctx context.Context, poolID resource.ID) ([]*agentToken, error)
	DeleteAgentToken(ctx context.Context, tokenID resource.ID) (*agentToken, error)
}

type webAuthorizer interface {
	CanAccess(context.Context, authz.Action, *authz.AccessRequest) bool
}

type (
	// templates may serialize hundreds of workspaces to JSON, so a struct with
	// only the fields needed is used rather than the full *workspace.Workspace
	// with its dozens of fields
	poolWorkspace struct {
		ID   resource.ID `json:"id"`
		Name string      `json:"name"`
	}

	poolWorkspaceList []poolWorkspace
)

// UnmarshalText is used by gorilla/schema to unmarshal a list of workspaces
func (l *poolWorkspaceList) UnmarshalText(v []byte) error {
	to := []poolWorkspace(*l)
	if err := json.Unmarshal(v, &to); err != nil {
		return err
	}
	*l = to
	return nil
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	// runners
	r.HandleFunc("/organizations/{organization_name}/runners", h.listRunners).Methods("GET")

	// agent pools
	r.HandleFunc("/organizations/{organization_name}/agent-pools", h.listAgentPools).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/agent-pools/create", h.createAgentPool).Methods("POST")
	r.HandleFunc("/agent-pools/{pool_id}", h.getAgentPool).Methods("GET")
	r.HandleFunc("/agent-pools/{pool_id}/update", h.updateAgentPool).Methods("POST")
	r.HandleFunc("/agent-pools/{pool_id}/delete", h.deleteAgentPool).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/pools", h.listAllowedPools).Methods("GET")

	// agent tokens
	r.HandleFunc("/agent-pools/{pool_id}/agent-tokens/create", h.createAgentToken).Methods("POST")
	r.HandleFunc("/agent-pools/{pool_id}/agent-tokens", h.listAgentTokens).Methods("GET")
	r.HandleFunc("/agent-tokens/{token_id}/delete", h.deleteAgentToken).Methods("POST")
}

// runner handlers

func (h *webHandlers) listRunners(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string `schema:"organization_name,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	serverRunners, err := h.svc.listServerRunners(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	agentRunners, err := h.svc.listRunnersByOrganization(r.Context(), params.Organization)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// order runners to show 'freshest' first
	runners := append(serverRunners, agentRunners...)
	slices.SortFunc(runners, func(a, b *RunnerMeta) int {
		if a.LastPingAt.Before(b.LastPingAt) {
			return 1
		} else {
			return -1
		}
	})

	h.Render("runners_list.tmpl", w, struct {
		organization.OrganizationPage
		html.Page[*RunnerMeta]
	}{
		OrganizationPage: organization.NewPage(r, "runners", params.Organization),
		Page: html.Page[*RunnerMeta]{
			Page:    resource.NewPage(runners, params.PageOptions, nil),
			Request: r,
		},
	})
}

// agent pool handlers

func (h *webHandlers) createAgentPool(w http.ResponseWriter, r *http.Request) {
	var opts CreateAgentPoolOptions
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pool, err := h.svc.CreateAgentPool(r.Context(), opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPool(pool.ID.String()), http.StatusFound)
}

func (h *webHandlers) updateAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// the client sends json-encoded slices of poolWorkspace structs whereas
	// updatePoolOptions expects slices of workspace IDs.
	//
	// also, the client handles two slices of workspaces, allowed-but-not-assigned,
	// and allowed-and-assigned, whereas updatePoolOptions handles them both as
	// a single slice of allowed workspaces.
	var params struct {
		Name                 string
		OrganizationScoped   bool              `schema:"organization_scoped"`
		AllowedButUnassigned poolWorkspaceList `schema:"allowed_workspaces"`
		AllowedAndAssigned   poolWorkspaceList `schema:"assigned_workspaces"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	opts := updatePoolOptions{
		Name:               &params.Name,
		OrganizationScoped: &params.OrganizationScoped,
		AllowedWorkspaces:  make([]resource.ID, len(params.AllowedButUnassigned)+len(params.AllowedAndAssigned)),
	}
	for i, allowed := range append(params.AllowedButUnassigned, params.AllowedAndAssigned...) {
		opts.AllowedWorkspaces[i] = allowed.ID
	}

	pool, err := h.svc.updateAgentPool(r.Context(), poolID, opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPool(pool.ID.String()), http.StatusFound)
}

func (h *webHandlers) listAgentPools(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string `schema:"organization_name,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pools, err := h.svc.listAgentPoolsByOrganization(r.Context(), params.Organization, listPoolOptions{})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_pools_list.tmpl", w, struct {
		organization.OrganizationPage
		html.Page[*Pool]
		Items []*Pool
	}{
		OrganizationPage: organization.NewPage(r, "agent pools", params.Organization),
		Page: html.Page[*Pool]{
			Page:    resource.NewPage(pools, params.PageOptions, nil),
			Request: r,
		},
		Items: pools,
	})
}

func (h *webHandlers) getAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pool, err := h.svc.GetAgentPool(r.Context(), poolID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// template requires three sets of workspaces:
	//
	// (a) assigned workspaces
	// (b) allowed but unassigned workspaces
	// (c) workspaces that are neither assigned nor allowed

	// fetch all workspaces in organization then distribute them among the three
	// sets documented above.
	allWorkspaces, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspacepkg.Workspace], error) {
		return h.workspaces.List(r.Context(), workspacepkg.ListOptions{
			PageOptions:  opts,
			Organization: &pool.Organization,
		})
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		assignedWorkspaces             = make([]poolWorkspace, 0, len(pool.AssignedWorkspaces))
		allowedButUnassignedWorkspaces = make([]poolWorkspace, 0, int(math.Abs(float64(len(pool.AllowedWorkspaces)-len(pool.AssignedWorkspaces)))))
		availableWorkspaces            = make([]poolWorkspace, 0, len(allWorkspaces)-len(allowedButUnassignedWorkspaces))
	)
	for _, ws := range allWorkspaces {
		isAssigned := slices.Contains(pool.AssignedWorkspaces, ws.ID)
		isAllowed := slices.Contains(pool.AllowedWorkspaces, ws.ID)
		if isAssigned {
			assignedWorkspaces = append(assignedWorkspaces, poolWorkspace{ID: ws.ID, Name: ws.Name})
		} else {
			if isAllowed {
				allowedButUnassignedWorkspaces = append(allowedButUnassignedWorkspaces, poolWorkspace{ID: ws.ID, Name: ws.Name})
			} else {
				availableWorkspaces = append(availableWorkspaces, poolWorkspace{ID: ws.ID, Name: ws.Name})
			}
		}
	}

	agents, err := h.svc.listRunnersByPool(r.Context(), poolID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_pool_get.tmpl", w, struct {
		organization.OrganizationPage
		Pool                           *Pool
		CanDeleteAgentPool             bool
		AllowedButUnassignedWorkspaces []poolWorkspace
		AssignedWorkspaces             []poolWorkspace
		AvailableWorkspaces            []poolWorkspace
		Agents                         []*RunnerMeta
	}{
		OrganizationPage:               organization.NewPage(r, pool.Name, pool.Organization),
		Pool:                           pool,
		CanDeleteAgentPool:             h.authorizer.CanAccess(r.Context(), authz.DeleteAgentPoolAction, &authz.AccessRequest{Organization: pool.Organization}),
		AllowedButUnassignedWorkspaces: allowedButUnassignedWorkspaces,
		AssignedWorkspaces:             assignedWorkspaces,
		AvailableWorkspaces:            availableWorkspaces,
		Agents:                         agents,
	})
}

func (h *webHandlers) deleteAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pool, err := h.svc.deleteAgentPool(r.Context(), poolID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "Deleted agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPools(pool.Organization), http.StatusFound)
}

func (h *webHandlers) listAllowedPools(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		WorkspaceID resource.ID  `schema:"workspace_id,required"`
		AgentPoolID *resource.ID `schema:"agent_pool_id"`
	}
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	ws, err := h.workspaces.Get(r.Context(), opts.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pools, err := h.svc.listAgentPoolsByOrganization(r.Context(), ws.Organization, listPoolOptions{
		AllowedWorkspaceID: &opts.WorkspaceID,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_pools_list_allowed.tmpl", w, struct {
		Pools         []*Pool
		CurrentPoolID *resource.ID
	}{
		Pools:         pools,
		CurrentPoolID: opts.AgentPoolID,
	})
}

// agent token handlers

func (h *webHandlers) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	var params struct {
		PoolID resource.ID `schema:"pool_id,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := h.svc.ListAgentTokens(r.Context(), params.PoolID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_token_list.tmpl", w, html.Page[*agentToken]{
		Page:    resource.NewPage(tokens, params.PageOptions, nil),
		Request: r,
	})
}

func (h *webHandlers) createAgentToken(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var opts CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, token, err := h.svc.CreateAgentToken(r.Context(), poolID, opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tokens.TokenFlashMessage(h, w, token); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.AgentPool(poolID.String()), http.StatusFound)
}

func (h *webHandlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("token_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	at, err := h.svc.DeleteAgentToken(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "Deleted token: "+at.Description)
	http.Redirect(w, r, paths.AgentPool(at.AgentPoolID.String()), http.StatusFound)
}
