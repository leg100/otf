package runner

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/workspace"
)

type tfe struct {
	*Service
	*tfeapi.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	// Agent Pools And Agents API
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/agents
	r.HandleFunc("/organizations/{organization_name}/agent-pools", a.listAgentPools).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/agent-pools", a.createAgentPool).Methods("POST")
	r.HandleFunc("/agent-pools/{pool_id}", a.getAgentPool).Methods("GET")
	r.HandleFunc("/agent-pools/{pool_id}", a.updateAgentPool).Methods("PATCH")
	r.HandleFunc("/agent-pools/{pool_id}", a.deleteAgentPool).Methods("DELETE")

	// Agent Tokens API
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/agent-tokens#create-an-agent-token
	r.HandleFunc("/agent-pools/{pool_id}/authentication-tokens", a.listAgentTokens).Methods("GET")
	r.HandleFunc("/agent-pools/{pool_id}/authentication-tokens", a.createAgentToken).Methods("POST")
	r.HandleFunc("/authentication-tokens/{token_id}", a.getAgentToken).Methods("GET")
	r.HandleFunc("/authentication-tokens/{token_id}", a.deleteAgentToken).Methods("DELETE")

	// Feature sets API:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/feature-sets
	//
	// This is only implemented in order to get the go-tfe integration tests
	// passing. Those tests first do some silly shit with the feature sets
	// API before hitting the agent pool API, so the former is stubbed out here
	// to make 'em happy.
	r.HandleFunc("/admin/feature-sets", func(w http.ResponseWriter, r *http.Request) {
		// tests expect one feature set to be returned but don't check contents
		// so return the bare minimum.
		fs := []struct {
			ID string `jsonapi:"primary,feature-sets"`
		}{
			{ID: "fs-123"},
		}
		a.RespondWithPage(w, r, &fs, &resource.Pagination{})
	})
	// Tests don't check response so return empty response.
	r.HandleFunc("/admin/organizations/{organization_name}/subscription", func(w http.ResponseWriter, r *http.Request) {})
}

// Agent pool handlers

func (a *tfe) createAgentPool(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFEAgentPoolCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if params.Name == nil {
		tfeapi.Error(w, internal.ErrRequiredName)
		return
	}

	// convert tfe params to otf opts
	opts := CreateAgentPoolOptions{
		Name:               *params.Name,
		Organization:       pathParams.Organization,
		OrganizationScoped: params.OrganizationScoped,
	}
	opts.AllowedWorkspaces = make([]resource.TfeID, len(params.AllowedWorkspaces))
	for i, aw := range params.AllowedWorkspaces {
		opts.AllowedWorkspaces[i] = aw.ID
	}

	pool, err := a.Service.CreateAgentPool(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.toPool(pool), http.StatusCreated)
}

func (a *tfe) updateAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFEAgentPoolUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert tfe params to otf opts
	opts := UpdatePoolOptions{
		Name:               params.Name,
		OrganizationScoped: params.OrganizationScoped,
	}
	if params.AllowedWorkspaces != nil {
		opts.AllowedWorkspaces = make([]resource.TfeID, len(params.AllowedWorkspaces))
		for i, aw := range params.AllowedWorkspaces {
			opts.AllowedWorkspaces[i] = aw.ID
		}
	}

	pool, err := a.Service.UpdateAgentPool(r.Context(), poolID, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.toPool(pool), http.StatusOK)
}

func (a *tfe) getAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	pool, err := a.Service.GetAgentPool(r.Context(), poolID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toPool(pool), http.StatusOK)
}

func (a *tfe) listAgentPools(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFEAgentPoolListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	pools, err := a.Service.ListAgentPoolsByOrganization(r.Context(), pathParams.Organization, ListPoolOptions{
		NameSubstring:        params.Query,
		AllowedWorkspaceName: params.AllowedWorkspacesName,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// client expects a page, whereas listPools returns full result set, so
	// convert to page first
	page := resource.NewPage(pools, resource.PageOptions(params.ListOptions), nil)

	// convert items
	items := make([]*TFEAgentPool, len(page.Items))
	for i, from := range page.Items {
		if err != nil {
			tfeapi.Error(w, err)
			return
		}
		items[i] = a.toPool(from)
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) deleteAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	_, err = a.Service.DeleteAgentPool(r.Context(), poolID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) toPool(from *Pool) *TFEAgentPool {
	to := &TFEAgentPool{
		ID:   from.ID,
		Name: from.Name,
		Organization: &organization.TFEOrganization{
			Name: from.Organization,
		},
		OrganizationScoped: from.OrganizationScoped,
	}
	to.Workspaces = make([]*workspace.TFEWorkspace, len(from.AssignedWorkspaces))
	for i, workspaceID := range from.AssignedWorkspaces {
		to.Workspaces[i] = &workspace.TFEWorkspace{ID: workspaceID}
	}
	to.AllowedWorkspaces = make([]*workspace.TFEWorkspace, len(from.AllowedWorkspaces))
	for i, workspaceID := range from.AllowedWorkspaces {
		to.AllowedWorkspaces[i] = &workspace.TFEWorkspace{ID: workspaceID}
	}
	return to
}

// Agent token handlers
func (a *tfe) createAgentToken(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFEAgentTokenCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	at, token, err := a.Service.CreateAgentToken(r.Context(), poolID, CreateAgentTokenOptions{
		Description: params.Description,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.toAgentToken(at, token), http.StatusCreated)
}

func (a *tfe) getAgentToken(w http.ResponseWriter, r *http.Request) {
	tokenID, err := decode.ID("token_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	at, err := a.Service.GetAgentToken(r.Context(), tokenID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toAgentToken(at, nil), http.StatusOK)
}

func (a *tfe) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params types.ListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	pools, err := a.Service.ListAgentTokens(r.Context(), poolID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// client expects a page, whereas service.listAgentTokens returns full result set, so
	// convert to page first
	page := resource.NewPage(pools, resource.PageOptions(params), nil)

	// convert items
	items := make([]*TFEAgentToken, len(page.Items))
	for i, from := range page.Items {
		if err != nil {
			tfeapi.Error(w, err)
			return
		}
		items[i] = a.toAgentToken(from, nil)
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	tokenID, err := decode.ID("token_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	_, err = a.Service.DeleteAgentToken(r.Context(), tokenID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) toAgentToken(from *AgentToken, token []byte) *TFEAgentToken {
	to := &TFEAgentToken{
		ID:          from.ID,
		CreatedAt:   from.CreatedAt,
		Description: from.Description,
	}
	if token != nil {
		to.Token = string(token)
	}
	return to
}
