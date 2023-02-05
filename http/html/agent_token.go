package html

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
)

func (app *Application) newAgentToken(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("agent_token_new.tmpl", w, r, org)
}

func (app *Application) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts otf.CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := app.CreateAgentToken(r.Context(), opts)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	//
	// TODO: replace with a helper func, 'flashTemplate'
	buf := new(bytes.Buffer)
	if err := app.renderTemplate("token_created.tmpl", buf, token.Token()); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.AgentTokens(opts.Organization), http.StatusFound)
}

func (app *Application) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := app.ListAgentTokens(r.Context(), organization)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
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

func (app *Application) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("agent_token_id", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	at, err := app.DeleteAgentToken(r.Context(), id)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	FlashSuccess(w, "Deleted token: "+at.Description())
	http.Redirect(w, r, paths.AgentTokens(at.Organization()), http.StatusFound)
}
