package ui

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/components"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runner"
	workspacepkg "github.com/leg100/otf/internal/workspace"
)

// runnersTableID is the CSS ID of the DOM element containing the table of
// runners.
const runnersTableID = "runners-table"

// runnerHandlers provides handlers for the web UI for runners
type runnerHandlers struct {
	svc                  runnerClient
	workspaces           *workspacepkg.Service
	logger               logr.Logger
	authorizer           authz.Interface
	websocketListHandler *components.WebsocketListHandler[*runner.RunnerMeta, *runner.RunnerEvent, runner.ListOptions]
}

// runnerClient gives web handlers access to the agents service endpoints
type runnerClient interface {
	CreateAgentPool(ctx context.Context, opts runner.CreateAgentPoolOptions) (*runner.Pool, error)
	UpdateAgentPool(ctx context.Context, poolID resource.TfeID, opts runner.UpdatePoolOptions) (*runner.Pool, error)
	ListAgentPoolsByOrganization(ctx context.Context, organization organization.Name, opts runner.ListPoolOptions) ([]*runner.Pool, error)
	GetAgentPool(ctx context.Context, poolID resource.TfeID) (*runner.Pool, error)
	DeleteAgentPool(ctx context.Context, poolID resource.TfeID) (*runner.Pool, error)

	Register(ctx context.Context, opts runner.RegisterRunnerOptions) (*runner.RunnerMeta, error)
	ListRunners(ctx context.Context, opts runner.ListOptions) ([]*runner.RunnerMeta, error)
	CreateAgentToken(ctx context.Context, poolID resource.TfeID, opts runner.CreateAgentTokenOptions) (*runner.AgentToken, []byte, error)
	GetAgentToken(ctx context.Context, tokenID resource.TfeID) (*runner.AgentToken, error)
	ListAgentTokens(ctx context.Context, poolID resource.TfeID) ([]*runner.AgentToken, error)
	DeleteAgentToken(ctx context.Context, tokenID resource.TfeID) (*runner.AgentToken, error)
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

func newRunnerHandlers(svc *runner.Service, opts runner.ServiceOptions) *runnerHandlers {
	return &runnerHandlers{
		authorizer: opts.Authorizer,
		logger:     opts.Logger,
		svc:        svc,
		workspaces: opts.WorkspaceService,
		websocketListHandler: &components.WebsocketListHandler[*runner.RunnerMeta, *runner.RunnerEvent, runner.ListOptions]{
			Logger:    opts.Logger,
			Client:    svc,
			Populator: runnersTable{},
			ID:        runnersTableID,
		},
	}
}

func (h *runnerHandlers) addHandlers(r *mux.Router) {
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

// runner handlers

func (h *runnerHandlers) listRunners(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		h.websocketListHandler.Handler(w, r)
		return
	}

	var params runner.ListOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	props := listRunnersProps{
		organization:      *params.Organization,
		hideServerRunners: params.HideServerRunners,
	}
	html.Render(listRunners(props), w, r)
}

// agent pool handlers

func (h *runnerHandlers) createAgentPool(w http.ResponseWriter, r *http.Request) {
	var opts runner.CreateAgentPoolOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	pool, err := h.svc.CreateAgentPool(r.Context(), opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "created agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPool(pool.ID), http.StatusFound)
}

func (h *runnerHandlers) updateAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
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

	pool, err := h.svc.UpdateAgentPool(r.Context(), poolID, opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "updated agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPool(pool.ID), http.StatusFound)
}

func (h *runnerHandlers) listAgentPools(w http.ResponseWriter, r *http.Request) {
	var params struct {
		resource.PageOptions
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	pools, err := h.svc.ListAgentPoolsByOrganization(r.Context(), params.Organization, runner.ListPoolOptions{})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	props := listAgentPoolProps{
		organization: params.Organization,
		pools:        resource.NewPage(pools, params.PageOptions, nil),
	}
	html.Render(listAgentPools(props), w, r)
}

func (h *runnerHandlers) getAgentPool(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		h.websocketListHandler.Handler(w, r)
		return
	}

	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	pool, err := h.svc.GetAgentPool(r.Context(), poolID)
	if err != nil {
		html.Error(r, w, err.Error())
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
		html.Error(r, w, err.Error())
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

	tokens, err := h.svc.ListAgentTokens(r.Context(), poolID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	agents, err := h.svc.ListRunners(r.Context(), runner.ListOptions{
		PoolID: &poolID,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	props := getAgentPoolProps{
		pool:                           pool,
		allowedButUnassignedWorkspaces: allowedButUnassignedWorkspaces,
		assignedWorkspaces:             assignedWorkspaces,
		availableWorkspaces:            availableWorkspaces,
		tokens:                         tokens,
		agents:                         resource.NewPage(agents, resource.PageOptions{}, nil),
		canDeleteAgentPool:             h.authorizer.CanAccess(r.Context(), authz.DeleteAgentPoolAction, pool.Organization),
	}
	html.Render(getAgentPool(props), w, r)
}

func (h *runnerHandlers) deleteAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	pool, err := h.svc.DeleteAgentPool(r.Context(), poolID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "Deleted agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPools(pool.Organization), http.StatusFound)
}

func (h *runnerHandlers) listAllowedPools(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		WorkspaceID resource.TfeID  `schema:"workspace_id,required"`
		AgentPoolID *resource.TfeID `schema:"agent_pool_id"`
	}
	if err := decode.All(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	ws, err := h.workspaces.Get(r.Context(), opts.WorkspaceID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	pools, err := h.svc.ListAgentPoolsByOrganization(r.Context(), ws.Organization, runner.ListPoolOptions{
		AllowedWorkspaceID: &opts.WorkspaceID,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	props := agentPoolListAllowedProps{
		pools:         pools,
		currentPoolID: opts.AgentPoolID,
	}
	html.Render(agentPoolListAllowed(props), w, r)
}

// agent token handlers

func (h *runnerHandlers) createAgentToken(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	var opts runner.CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	_, token, err := h.svc.CreateAgentToken(r.Context(), poolID, opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	if err := components.TokenFlashMessage(w, token); err != nil {
		html.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.AgentPool(poolID), http.StatusFound)
}

func (h *runnerHandlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("token_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	at, err := h.svc.DeleteAgentToken(r.Context(), id)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "Deleted token: "+at.Description)
	http.Redirect(w, r, paths.AgentPool(at.AgentPoolID), http.StatusFound)
}
