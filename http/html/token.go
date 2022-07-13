package html

import (
	"bytes"
	"net/http"
	"sort"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

// tokenList exposes a list of tokens to a template
type tokenList struct {
	// list template expects pagination object but we don't paginate token
	// listing
	*otf.Pagination
	Items []*otf.Token
}

func (app *Application) newTokenHandler(w http.ResponseWriter, r *http.Request) {
	app.render("token_new.tmpl", w, r, nil)
}

func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var opts otf.TokenCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	token, err := app.UserService().CreateToken(r.Context(), user, &opts)
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

	http.Redirect(w, r, listTokenPath(), http.StatusFound)
}

func (app *Application) tokensHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// display tokens by creation date, newest first
	sort.Slice(user.Tokens, func(i, j int) bool {
		return user.Tokens[i].CreatedAt().After(user.Tokens[j].CreatedAt())
	})
	app.render("token_list.tmpl", w, r, tokenList{
		Pagination: &otf.Pagination{},
		Items:      user.Tokens,
	})
}

func (app *Application) deleteTokenHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "missing id", http.StatusUnprocessableEntity)
		return
	}
	if err := app.UserService().DeleteToken(r.Context(), user, id); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "Deleted token")
	http.Redirect(w, r, listTokenPath(), http.StatusFound)
}
