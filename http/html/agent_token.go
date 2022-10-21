package html

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

func (app *Application) newAgentToken(w http.ResponseWriter, r *http.Request) {
	app.render("agent_token_new.tmpl", w, r, organizationRequest{r})
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
	tokens, err := app.ListAgentTokens(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// tokenList exposes a list of tokens to a template
	type tokenList struct {
		// list template expects pagination object but we don't paginate token
		// listing
		*otf.Pagination
		Items []*otf.AgentToken
		organizationRoute
	}
	app.render("agent_token_list.tmpl", w, r, tokenList{
		Pagination:        &otf.Pagination{},
		Items:             tokens,
		organizationRoute: organizationRequest{r},
	})
}

func (app *Application) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, listAgentTokenPath(organizationRequest{r}), http.StatusFound)
}
