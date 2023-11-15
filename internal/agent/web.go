package agent

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	html.Renderer
	svc Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	// agent pools
	r.HandleFunc("/organizations/{organization_name}/agent-pools", h.listAgentPools).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/agent-pools/create", h.createAgentPool).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/agent-pools/new", h.newAgentToken).Methods("GET")
	r.HandleFunc("/agent-pools/{pool_id}", h.getAgentPool).Methods("GET")
	r.HandleFunc("/agent-pools/{pool_id}/update", h.updateAgentPool).Methods("POST")
	r.HandleFunc("/agent-pools/{pool_id}/delete", h.deleteAgentPool).Methods("POST")

	// agent tokens
	r.HandleFunc("/organizations/{organization_name}/agent-tokens", h.listAgentTokens).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/create", h.createAgentToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/new", h.newAgentToken).Methods("GET")
	r.HandleFunc("/agent-tokens/{agent_token_id}/delete", h.deleteAgentToken).Methods("POST")
}

//
// Agent pool handlers
//

func (h *webHandlers) newAgentPool(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("agent_token_new.tmpl", w, struct {
		organization.OrganizationPage
	}{
		OrganizationPage: organization.NewPage(r, "new agent token", org),
	})
}

func (h *webHandlers) createAgentPool(w http.ResponseWriter, r *http.Request) {
	var opts createAgentPoolOptions
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pool, err := h.svc.createAgentPool(r.Context(), opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPools(pool.ID), http.StatusFound)
}

func (h *webHandlers) updateAgentPool(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.Param("pool_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var opts updatePoolOptions
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pool, err := h.svc.updateAgentPool(r.Context(), poolID, opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated agent pool: "+pool.Name)
	http.Redirect(w, r, paths.AgentPools(pool.ID), http.StatusFound)
}

func (h *webHandlers) listAgentPools(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pools, err := h.svc.listAgentPools(r.Context(), listPoolOptions{
		Organization: &org,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_pool_list.tmpl", w, struct {
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

	pool, err := h.svc.getAgentPool(r.Context(), poolID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_pool_get.tmpl", w, struct {
		organization.OrganizationPage
		Pool *Pool
	}{
		OrganizationPage: organization.NewPage(r, "agent pools", pool.Name),
		Pool:             pool,
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

//
// Agent token handlers
//

func (h *webHandlers) newAgentToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("agent_token_new.tmpl", w, struct {
		organization.OrganizationPage
	}{
		OrganizationPage: organization.NewPage(r, "new agent token", org),
	})
}

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
	http.Redirect(w, r, paths.AgentTokens(poolID), http.StatusFound)
}

func (h *webHandlers) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := h.svc.ListAgentTokens(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_token_list.tmpl", w, struct {
		organization.OrganizationPage
		// list template expects pagination object but we don't paginate token
		// listing
		*resource.Pagination
		Items []*agentToken
	}{
		OrganizationPage: organization.NewPage(r, "agent tokens", org),
		Pagination:       &resource.Pagination{},
		Items:            tokens,
	})
}

func (h *webHandlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("agent_token_id", r)
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
	http.Redirect(w, r, paths.AgentTokens(""), http.StatusFound)
}
