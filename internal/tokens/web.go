package tokens

import (
	"bytes"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	html.Renderer

	svc       TokensService
	siteToken string
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	//
	// Unauthenticated routes
	//
	r.HandleFunc("/admin/login", h.adminLogin).Methods("POST")

	//
	// Authenticated routes
	//
	r = html.UIRouter(r)

	// user tokens
	r.HandleFunc("/profile/tokens", h.userTokens).Methods("GET")
	r.HandleFunc("/profile/tokens/delete", h.deleteUserToken).Methods("POST")
	r.HandleFunc("/profile/tokens/new", h.newUserToken).Methods("GET")
	r.HandleFunc("/profile/tokens/create", h.createUserToken).Methods("POST")

	// agent tokens
	r.HandleFunc("/organizations/{organization_name}/agent-tokens", h.listAgentTokens).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/create", h.createAgentToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/new", h.newAgentToken).Methods("GET")
	r.HandleFunc("/agent-tokens/{agent_token_id}/delete", h.deleteAgentToken).Methods("POST")

	r.HandleFunc("/logout", h.logout).Methods("POST")

	// terraform login opens a browser to this hardcoded URL
	r.HandleFunc("/settings/tokens", h.userTokens).Methods("GET")
}

func (h *webHandlers) newUserToken(w http.ResponseWriter, r *http.Request) {
	h.Render("token_new.tmpl", w, html.NewSitePage(r, "new user token"))
}

func (h *webHandlers) createUserToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateUserTokenOptions
	if err := decode.Form(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, token, err := h.svc.CreateUserToken(r.Context(), opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.tokenFlashMessage(w, token); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}

func (h *webHandlers) userTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.svc.ListUserTokens(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// re-order tokens by creation date, newest first
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].CreatedAt.After(tokens[j].CreatedAt)
	})

	h.Render("user_token_list.tmpl", w, struct {
		html.SitePage
		// list template expects pagination object but we don't paginate token
		// listing
		*internal.Pagination
		Items []*UserToken
	}{
		SitePage:   html.NewSitePage(r, "user tokens"),
		Pagination: &internal.Pagination{},
		Items:      tokens,
	})
}

func (h *webHandlers) deleteUserToken(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		h.Error(w, "missing id", http.StatusUnprocessableEntity)
		return
	}
	if err := h.svc.DeleteUserToken(r.Context(), id); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Deleted token")
	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}

//
// Site Admin handlers
//

// adminLogin logs in a site admin
func (h *webHandlers) adminLogin(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if token != h.siteToken {
		html.FlashError(w, "incorrect token")
		http.Redirect(w, r, paths.AdminLogin(), http.StatusFound)
		return
	}

	err = h.svc.StartSession(w, r, StartSessionOptions{
		Username: internal.String(auth.SiteAdminUsername),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//
// Agent token handlers
//

func (h *webHandlers) newAgentToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("agent_token_new.tmpl", w, struct {
		organization.OrganizationPage
	}{
		OrganizationPage: organization.NewPage(r, "new agent token", org),
	})
}

func (h *webHandlers) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := h.svc.CreateAgentToken(r.Context(), opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.tokenFlashMessage(w, token); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.AgentTokens(opts.Organization), http.StatusFound)
}

func (h *webHandlers) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := h.svc.ListAgentTokens(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("agent_token_list.tmpl", w, struct {
		organization.OrganizationPage
		// list template expects pagination object but we don't paginate token
		// listing
		*internal.Pagination
		Items []*AgentToken
	}{
		OrganizationPage: organization.NewPage(r, "agent tokens", org),
		Pagination:       &internal.Pagination{},
		Items:            tokens,
	})
}

func (h *webHandlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("agent_token_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	at, err := h.svc.DeleteAgentToken(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "Deleted token: "+at.Description)
	http.Redirect(w, r, paths.AgentTokens(at.Organization), http.StatusFound)
}

func (h *webHandlers) logout(w http.ResponseWriter, r *http.Request) {
	html.SetCookie(w, sessionCookie, "", &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *webHandlers) tokenFlashMessage(w http.ResponseWriter, token []byte) error {
	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := h.RenderTemplate("token_created.tmpl", buf, string(token)); err != nil {
		return err
	}
	html.FlashSuccess(w, buf.String())
	return nil
}
