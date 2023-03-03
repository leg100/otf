package auth

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

func (h *web) addAgentTokenHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/agent-tokens", h.listAgentTokens).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/create", h.createAgentToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/new", h.newAgentToken).Methods("GET")
	r.HandleFunc("/agent-tokens/{agent_token_id}/delete", h.deleteAgentToken).Methods("POST")
}

func (h *web) newAgentToken(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("agent_token_new.tmpl", w, r, organization)
}

func (h *web) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts otf.CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := h.svc.createAgentToken(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := h.RenderTemplate("token_created.tmpl", buf, token.Token); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.AgentTokens(opts.Organization), http.StatusFound)
}

func (app *web) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := app.svc.listAgentTokens(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("agent_token_list.tmpl", w, r, struct {
		// list template expects pagination object but we don't paginate token
		// listing
		*otf.Pagination
		Items        []*otf.AgentToken
		Organization string
	}{
		Pagination:   &otf.Pagination{},
		Items:        tokens,
		Organization: organization,
	})
}

func (app *web) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("agent_token_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	at, err := app.svc.deleteAgentToken(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "Deleted token: "+at.Description)
	http.Redirect(w, r, paths.AgentTokens(at.Organization), http.StatusFound)
}
