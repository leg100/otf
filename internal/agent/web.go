package agent

import (
	"encoding/json"
	"math"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
	workspacepkg "github.com/leg100/otf/internal/workspace"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	html.Renderer
	svc              Service
	workspaceService workspacepkg.WorkspaceService
	logger           logr.Logger
}

type (
	// templates may serialize hundreds of workspaces to JSON, so a struct with
	// only the fields needed is used rather than the full *workspace.Workspace
	// with its dozens of fields
	poolWorkspace struct {
		ID   string `json:"id"`
		Name string `json:"name"`
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

	// agents
	r.HandleFunc("/organizations/{organization_name}/agents", h.listAgents).Methods("GET")

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

// agent handlers

func (h *webHandlers) listAgents(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	serverAgents, err := h.svc.listServerAgents(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	poolAgents, err := h.svc.listAgentsByOrganization(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// order agents to show 'freshest' at the top
	agents := append(serverAgents, poolAgents...)
	slices.SortFunc(agents, func(a, b *Agent) int {
		if a.LastPingAt.Before(b.LastPingAt) {
			return 1
		} else {
			return -1
		}
	})

	h.Render("agents_list.tmpl", w, struct {
		organization.OrganizationPage
		Agents []*Agent
	}{
		OrganizationPage: organization.NewPage(r, "agents", org),
		Agents:           agents,
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
	http.Redirect(w, r, paths.AgentPool(pool.ID), http.StatusFound)
}

func (h *webHandlers) updateAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.Param("pool_id", r)
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
		AllowedWorkspaces:  make([]string, len(params.AllowedButUnassigned)+len(params.AllowedAndAssigned)),
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
	http.Redirect(w, r, paths.AgentPool(pool.ID), http.StatusFound)
}

func (h *webHandlers) listAgentPools(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pools, err := h.svc.listAgentPoolsByOrganization(r.Context(), org, listPoolOptions{})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_pools_list.tmpl", w, struct {
		organization.OrganizationPage
		// list template expects pagination object but we don't paginate token
		// listing
		*resource.Pagination
		Items []*Pool
	}{
		OrganizationPage: organization.NewPage(r, "agent pools", org),
		Pagination:       &resource.Pagination{},
		Items:            pools,
	})
}

func (h *webHandlers) getAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.Param("pool_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pool, err := h.svc.GetAgentPool(r.Context(), poolID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	subject, err := internal.SubjectFromContext(r.Context())
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
		return h.workspaceService.ListWorkspaces(r.Context(), workspacepkg.ListOptions{
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

	tokens, err := h.svc.ListAgentTokens(r.Context(), poolID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	agents, err := h.svc.listAgentsByPool(r.Context(), poolID)
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
		Tokens                         []*agentToken
		Agents                         []*Agent
	}{
		OrganizationPage:               organization.NewPage(r, pool.Name, pool.Organization),
		Pool:                           pool,
		CanDeleteAgentPool:             subject.CanAccessOrganization(rbac.DeleteAgentPoolAction, pool.Organization),
		AllowedButUnassignedWorkspaces: allowedButUnassignedWorkspaces,
		AssignedWorkspaces:             assignedWorkspaces,
		AvailableWorkspaces:            availableWorkspaces,
		Tokens:                         tokens,
		Agents:                         agents,
	})
}

func (h *webHandlers) deleteAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.Param("pool_id", r)
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
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.workspaceService.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pools, err := h.svc.listAgentPoolsByOrganization(r.Context(), ws.Organization, listPoolOptions{
		AllowedWorkspaceID: &workspaceID,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_pools_list_allowed.tmpl", w, pools)
}

// agent token handlers

func (h *webHandlers) createAgentToken(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.Param("pool_id", r)
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
	http.Redirect(w, r, paths.AgentPool(poolID), http.StatusFound)
}

func (h *webHandlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("token_id", r)
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
	http.Redirect(w, r, paths.AgentPool(at.AgentPoolID), http.StatusFound)
}
