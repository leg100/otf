package auth

import (
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	html.Renderer

	svc AuthService
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	// Unauthenticated routes
	r.HandleFunc("/admin/login", h.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", h.adminLogin).Methods("POST")

	// Authenticated routes
	r = html.UIRouter(r)

	r.HandleFunc("/logout", h.logout).Methods("POST")

	r.HandleFunc("/organizations/{name}/users", h.listOrganizationUsers).Methods("GET")
	r.HandleFunc("/profile", h.profileHandler).Methods("GET")
	r.HandleFunc("/admin", h.site).Methods("GET")

	// user tokens
	r.HandleFunc("/profile/tokens", h.userTokens).Methods("GET")
	r.HandleFunc("/profile/tokens/delete", h.deleteUserToken).Methods("POST")
	r.HandleFunc("/profile/tokens/new", h.newUserToken).Methods("GET")
	r.HandleFunc("/profile/tokens/create", h.createUserToken).Methods("POST")

	// terraform login opens a browser to this hardcoded URL
	r.HandleFunc("/settings/tokens", h.userTokens).Methods("GET")

	// team pages
	h.addTeamHandlers(r)
}

func (h *webHandlers) logout(w http.ResponseWriter, r *http.Request) {
	html.SetCookie(w, sessionCookie, "", &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *webHandlers) listOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	users, err := h.svc.ListOrganizationUsers(r.Context(), name)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("users_list.tmpl", w, struct {
		html.SitePage
		Users []*User
	}{
		SitePage: html.NewSitePage(r, "users"),
		Users:    users,
	})
}

func (h *webHandlers) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("profile.tmpl", w, struct {
		html.SitePage
		User internal.Subject
	}{
		SitePage: html.NewSitePage(r, "profile"),
		User:     user,
	})
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (h *webHandlers) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	h.Render("site_admin_login.tmpl", w, html.NewSitePage(r, "site admin login"))
}

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
		Username: internal.String(SiteAdminUsername),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *webHandlers) site(w http.ResponseWriter, r *http.Request) {
	user, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("site.tmpl", w, struct {
		html.SitePage
		User internal.Subject
	}{
		SitePage: html.NewSitePage(r, "site"),
		User:     user,
	})
}

//
// User tokens
//

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

	if err := tokens.TokenFlashMessage(h, w, token); err != nil {
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
		*resource.Pagination
		Items []*UserToken
	}{
		SitePage:   html.NewSitePage(r, "user tokens"),
		Pagination: &resource.Pagination{},
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
