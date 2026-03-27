package ui

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/workspace"
)

type Handlers struct {
	Client     Client
	Authorizer authz.Interface
}

type Client interface {
	ListRunners(ctx context.Context, opts runner.ListOptions) ([]*runner.RunnerMeta, error)
	CreateAgentPool(ctx context.Context, opts runner.CreateAgentPoolOptions) (*runner.Pool, error)
	UpdateAgentPool(ctx context.Context, poolID resource.TfeID, opts runner.UpdatePoolOptions) (*runner.Pool, error)
	ListAgentPoolsByOrganization(ctx context.Context, organization organization.Name, opts runner.ListPoolOptions) ([]*runner.Pool, error)
	GetAgentPool(ctx context.Context, poolID resource.TfeID) (*runner.Pool, error)
	DeleteAgentPool(ctx context.Context, poolID resource.TfeID) (*runner.Pool, error)
	CreateAgentToken(ctx context.Context, poolID resource.TfeID, opts runner.CreateAgentTokenOptions) (*runner.AgentToken, []byte, error)
	ListAgentTokens(ctx context.Context, poolID resource.TfeID) ([]*runner.AgentToken, error)
	DeleteAgentToken(ctx context.Context, tokenID resource.TfeID) (*runner.AgentToken, error)
	GetWorkspace(context.Context, resource.TfeID) (*workspace.Workspace, error)
	ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)
}

func (h *Handlers) AddHandlers(r *mux.Router) {
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
	r.HandleFunc("/agent-tokens/{token_id}/delete", h.deleteAgentToken).Methods("POST")
}

type (
	// templates may serialize hundreds of workspaces to JSON, so a struct with
	// only the fields needed is used rather than the full *workspace.Workspace
	// with its dozens of fields
	poolWorkspace struct {
		ID   resource.TfeID `json:"id"`
		Name string         `json:"name"`
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

// runner handlers

func (h *Handlers) listRunners(w http.ResponseWriter, r *http.Request) {
	var params runner.ListOptions
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	runners, err := h.Client.ListRunners(r.Context(), params)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := listRunnersProps{
		organization:      *params.Organization,
		hideServerRunners: params.HideServerRunners,
		page:              resource.NewPage(runners, resource.PageOptions{}, nil),
	}
	helpers.RenderPage(
		listRunners(props),
		"runners",
		w,
		r,
		helpers.WithOrganization(*params.Organization),
		helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "Runners"}),
	)
}

// agent pool handlers

func (h *Handlers) createAgentPool(w http.ResponseWriter, r *http.Request) {
	var opts runner.CreateAgentPoolOptions
	if err := decode.All(&opts, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	pool, err := h.Client.CreateAgentPool(r.Context(), opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "created agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPool(pool.ID), http.StatusFound)
}

func (h *Handlers) updateAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	opts := runner.UpdatePoolOptions{
		Name:               &params.Name,
		OrganizationScoped: &params.OrganizationScoped,
		AllowedWorkspaces:  make([]resource.TfeID, len(params.AllowedButUnassigned)+len(params.AllowedAndAssigned)),
	}
	for i, allowed := range append(params.AllowedButUnassigned, params.AllowedAndAssigned...) {
		opts.AllowedWorkspaces[i] = allowed.ID
	}

	pool, err := h.Client.UpdateAgentPool(r.Context(), poolID, opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "updated agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPool(pool.ID), http.StatusFound)
}

func (h *Handlers) listAgentPools(w http.ResponseWriter, r *http.Request) {
	var params struct {
		resource.PageOptions
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	pools, err := h.Client.ListAgentPoolsByOrganization(r.Context(), params.Organization, runner.ListPoolOptions{})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := listAgentPoolProps{
		organization: params.Organization,
		pools:        resource.NewPage(pools, params.PageOptions, nil),
	}
	helpers.RenderPage(
		listAgentPools(props),
		"agent pools",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "Agent Pools"}),
	)
}

func (h *Handlers) getAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	pool, err := h.Client.GetAgentPool(r.Context(), poolID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// template requires three sets of workspaces:
	//
	// (a) assigned workspaces
	// (b) allowed but unassigned workspaces
	// (c) workspaces that are neither assigned nor allowed

	// fetch all workspaces in organization then distribute them among the three
	// sets documented above.
	allWorkspaces, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Workspace], error) {
		return h.Client.ListWorkspaces(r.Context(), workspace.ListOptions{
			PageOptions:  opts,
			Organization: &pool.Organization,
		})
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
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

	tokens, err := h.Client.ListAgentTokens(r.Context(), poolID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	agents, err := h.Client.ListRunners(r.Context(), runner.ListOptions{
		PoolID: &poolID,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := getAgentPoolProps{
		pool:                           pool,
		allowedButUnassignedWorkspaces: allowedButUnassignedWorkspaces,
		assignedWorkspaces:             assignedWorkspaces,
		availableWorkspaces:            availableWorkspaces,
		tokens:                         tokens,
		agents:                         resource.NewPage(agents, resource.PageOptions{}, nil),
		canDeleteAgentPool:             h.Authorizer.CanAccess(r.Context(), authz.DeleteAgentPoolAction, pool.Organization),
	}
	helpers.RenderPage(
		getAgentPool(props),
		pool.Name,
		w,
		r,
		helpers.WithOrganization(pool.Organization),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Agent Pools", Link: paths.AgentPools(props.pool.Organization)},
			helpers.Breadcrumb{Name: props.pool.Name},
		),
	)
}

func (h *Handlers) deleteAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	pool, err := h.Client.DeleteAgentPool(r.Context(), poolID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "Deleted agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPools(pool.Organization), http.StatusFound)
}

func (h *Handlers) listAllowedPools(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		WorkspaceID resource.TfeID  `schema:"workspace_id,required"`
		AgentPoolID *resource.TfeID `schema:"agent_pool_id"`
	}
	if err := decode.All(&opts, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	ws, err := h.Client.GetWorkspace(r.Context(), opts.WorkspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	pools, err := h.Client.ListAgentPoolsByOrganization(r.Context(), ws.Organization, runner.ListPoolOptions{
		AllowedWorkspaceID: &opts.WorkspaceID,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := agentPoolListAllowedProps{
		pools:         pools,
		currentPoolID: opts.AgentPoolID,
	}
	helpers.Render(agentPoolListAllowed(props), w, r)
}

// agent token handlers

func (h *Handlers) createAgentToken(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	var opts runner.CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	_, token, err := h.Client.CreateAgentToken(r.Context(), poolID, opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	if err := helpers.TokenFlashMessage(w, token); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.AgentPool(poolID), http.StatusFound)
}

func (h *Handlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("token_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	at, err := h.Client.DeleteAgentToken(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "Deleted token: "+at.Description)
	http.Redirect(w, r, paths.AgentPool(at.AgentPoolID), http.StatusFound)
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
