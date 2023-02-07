package agenttoken

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

// webApp is the web application for organizations
type htmlHandlers struct {
	otf.Renderer // renders templates

	app appService // provide access to org service
}

func (app *htmlHandlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/agent-tokens", app.listAgentTokens)
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/create", app.createAgentToken)
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/new", app.newAgentToken)
	r.HandleFunc("/agent-tokens/{agent_token_id}/delete", app.deleteAgentToken)
}

func (app *htmlHandlers) newAgentToken(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("agent_token_new.tmpl", w, r, org)
}

func (app *htmlHandlers) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts otf.CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := app.CreateAgentToken(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	//
	// TODO: replace with a helper func, 'flashTemplate'
	buf := new(bytes.Buffer)
	if err := app.renderTemplate("token_created.tmpl", buf, token.Token()); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.AgentTokens(opts.Organization), http.StatusFound)
}

func (app *htmlHandlers) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := app.ListAgentTokens(r.Context(), organization)
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

func (app *htmlHandlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("agent_token_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	at, err := app.DeleteAgentToken(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "Deleted token: "+at.Description())
	http.Redirect(w, r, paths.AgentTokens(at.Organization()), http.StatusFound)
}
