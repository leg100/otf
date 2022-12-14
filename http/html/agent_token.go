package html

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
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
	opts := otf.AgentTokenCreateOptions{
		Description:      r.FormValue("description"),
		OrganizationName: mux.Vars(r)["organization_name"],
	}
	token, err := app.CreateAgentToken(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := app.renderTemplate("token_created.tmpl", buf, token.Token()); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, buf.String())

	http.Redirect(w, r, listAgentTokenPath(token), http.StatusFound)
}

func (app *Application) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tokens, err := app.ListAgentTokens(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("agent_token_list.tmpl", w, r, struct {
		// list template expects pagination object but we don't paginate token
		// listing
		*otf.Pagination
		Items []*otf.AgentToken
		*otf.Organization
	}{
		Pagination:   &otf.Pagination{},
		Items:        tokens,
		Organization: org,
	})
}

func (app *Application) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "missing id", http.StatusUnprocessableEntity)
		return
	}
	if err := app.DeleteAgentToken(r.Context(), id, mux.Vars(r)["organization_name"]); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "Deleted token")
	http.Redirect(w, r, listAgentTokenPath(org), http.StatusFound)
}
