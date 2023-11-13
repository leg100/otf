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
	Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	// agent tokens
	r.HandleFunc("/organizations/{organization_name}/agent-tokens", h.listAgentTokens).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/create", h.createAgentToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/new", h.newAgentToken).Methods("GET")
	r.HandleFunc("/agent-tokens/{agent_token_id}/delete", h.deleteAgentToken).Methods("POST")
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
	var opts CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := h.CreateAgentToken(r.Context(), opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tokens.TokenFlashMessage(h, w, token); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.AgentTokens(opts.Organization), http.StatusFound)
}

func (h *webHandlers) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := h.ListAgentTokens(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_token_list.tmpl", w, struct {
		organization.OrganizationPage
		// list template expects pagination object but we don't paginate token
		// listing
		*resource.Pagination
		Items []*AgentToken
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

	at, err := h.DeleteAgentToken(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "Deleted token: "+at.Description)
	http.Redirect(w, r, paths.AgentTokens(at.Organization), http.StatusFound)
}
