package html

import (
	"bytes"
	"net/http"
	"sort"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
)

// tokenList exposes a list of tokens to a template
type tokenList struct {
	// list template expects pagination object but we don't paginate token
	// listing
	*otf.Pagination
	Items []*otf.Token
}

func (app *Application) newTokenHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("token_new.tmpl", w, r, nil)
}

func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var opts otf.TokenCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	token, err := app.CreateToken(r.Context(), user.ID(), &opts)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := app.renderTemplate("token_created.tmpl", buf, token.Token()); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}

func (app *Application) tokensHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tokens, err := app.ListTokens(r.Context(), user.ID())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// re-order tokens by creation date, newest first
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].CreatedAt().After(tokens[j].CreatedAt())
	})

	app.Render("token_list.tmpl", w, r, tokenList{
		Pagination: &otf.Pagination{},
		Items:      tokens,
	})
}

func (app *Application) deleteTokenHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		Error(w, "missing id", http.StatusUnprocessableEntity)
		return
	}
	if err := app.DeleteToken(r.Context(), user.ID(), id); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	FlashSuccess(w, "Deleted token")
	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}
