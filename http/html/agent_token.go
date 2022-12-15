package html

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (app *Application) newAgentToken(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("agent_token_new.tmpl", w, r, org)
}

func (app *Application) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts otf.CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := app.CreateAgentToken(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	//
	// TODO: replace with a helper func, 'flashTemplate'
	buf := new(bytes.Buffer)
	if err := app.renderTemplate("token_created.tmpl", buf, token.Token()); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, buf.String())

	http.Redirect(w, r, agentTokensPath(opts.OrganizationName), http.StatusFound)
}

func (app *Application) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := app.ListAgentTokens(r.Context(), organization)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("agent_token_list.tmpl", w, r, struct {
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
	type parameters struct {
		Organization string `schema:"organization_name,required"`
		ID           string `schema:"id,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := app.DeleteAgentToken(r.Context(), params.ID, params.Organization); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "Deleted token")
	http.Redirect(w, r, agentTokensPath(params.Organization), http.StatusFound)
}
