package tokens

import (
	"bytes"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	otf.Renderer

	svc       TokensService
	siteToken string
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	//
	// Unauthenticated routes
	//
	r.HandleFunc("/admin/login", h.adminLoginHandler).Methods("POST")

	//
	// Authenticated routes
	//
	r = html.UIRouter(r)

	// user tokens
	r.HandleFunc("/profile/tokens", h.tokensHandler).Methods("GET")
	r.HandleFunc("/profile/tokens/delete", h.deleteTokenHandler).Methods("POST")
	r.HandleFunc("/profile/tokens/new", h.newTokenHandler).Methods("GET")
	r.HandleFunc("/profile/tokens/create", h.createTokenHandler).Methods("POST")

	// agent tokens
	r.HandleFunc("/organizations/{organization_name}/agent-tokens", h.listAgentTokens).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/create", h.createAgentToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/new", h.newAgentToken).Methods("GET")
	r.HandleFunc("/agent-tokens/{agent_token_id}/delete", h.deleteAgentToken).Methods("POST")

	r.HandleFunc("/logout", h.logoutHandler).Methods("POST")

	// terraform login opens a browser to this hardcoded URL
	r.HandleFunc("/settings/tokens", h.tokensHandler).Methods("GET")
}

func (h *webHandlers) newTokenHandler(w http.ResponseWriter, r *http.Request) {
	h.Render("token_new.tmpl", w, r, nil)
}

func (h *webHandlers) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var opts CreateTokenOptions
	if err := decode.Form(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, token, err := h.svc.CreateToken(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := h.RenderTemplate("token_created.tmpl", buf, string(token)); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}

func (h *webHandlers) tokensHandler(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.svc.ListTokens(r.Context())
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
	id := r.FormValue("id")
	if id == "" {
		html.Error(w, "missing id", http.StatusUnprocessableEntity)
		return
	}
	if err := h.svc.DeleteToken(r.Context(), id); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Deleted token")
	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}

//
// Site Admin handlers
//

// adminLoginHandler logs in a site admin
func (h *webHandlers) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if token != h.siteToken {
		html.FlashError(w, "incorrect token")
		http.Redirect(w, r, paths.AdminLogin(), http.StatusFound)
		return
	}

	err = h.svc.StartSession(w, r, StartSessionOptions{
		Username: otf.String(auth.SiteAdminUsername),
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//
// Agent token handlers
//

func (h *webHandlers) newAgentToken(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("agent_token_new.tmpl", w, r, organization)
}

func (h *webHandlers) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := h.svc.CreateAgentToken(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := h.RenderTemplate("token_created.tmpl", buf, token); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.AgentTokens(opts.Organization), http.StatusFound)
}

func (h *webHandlers) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := h.svc.ListAgentTokens(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_token_list.tmpl", w, r, struct {
		// list template expects pagination object but we don't paginate token
		// listing
		*otf.Pagination
		Items        []*AgentToken
		Organization string
	}{
		Pagination:   &otf.Pagination{},
		Items:        tokens,
		Organization: organization,
	})
}

func (h *webHandlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("agent_token_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	at, err := h.svc.DeleteAgentToken(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "Deleted token: "+at.Description)
	http.Redirect(w, r, paths.AgentTokens(at.Organization), http.StatusFound)
}

func (h *webHandlers) logoutHandler(w http.ResponseWriter, r *http.Request) {
	html.SetCookie(w, sessionCookie, "", &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}
