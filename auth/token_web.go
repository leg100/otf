package auth

import (
	"bytes"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

func (h *webHandlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/profile/tokens", h.tokensHandler).Methods("GET")
	r.HandleFunc("/profile/tokens/delete", h.deleteTokenHandler).Methods("POST")
	r.HandleFunc("/profile/tokens/new", h.newTokenHandler).Methods("GET")
	r.HandleFunc("/profile/tokens/create", h.createTokenHandler).Methods("POST")

	// terraform login opens a browser to this hardcoded URL
	r.HandleFunc("/app/settings/tokens", h.tokensHandler).Methods("GET")
}

func (h *webHandlers) newTokenHandler(w http.ResponseWriter, r *http.Request) {
	h.Render("token_new.tmpl", w, r, nil)
}

func (h *webHandlers) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var opts TokenCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	token, err := h.svc.CreateToken(r.Context(), user.ID, &opts)
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

	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}

func (h *webHandlers) tokensHandler(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tokens, err := h.svc.ListTokens(r.Context(), user.ID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// re-order tokens by creation date, newest first
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].CreatedAt.After(tokens[j].CreatedAt)
	})

	h.Render("token_list.tmpl", w, r, struct {
		// list template expects pagination object but we don't paginate token
		// listing
		*otf.Pagination
		Items []*Token
	}{
		Pagination: &otf.Pagination{},
		Items:      tokens,
	})
}

func (h *webHandlers) deleteTokenHandler(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		html.Error(w, "missing id", http.StatusUnprocessableEntity)
		return
	}
	if err := h.svc.DeleteToken(r.Context(), user.ID, id); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Deleted token")
	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}
