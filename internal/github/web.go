package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/user"
)

const (
	// GithubPath is the URL path for the endpoint receiving VCS events from the
	// Github App
	AppEventsPath = "/webhooks/github-app"
)

type webHandlers struct {
	html.Renderer
	internal.HostnameService

	svc            Service
	GithubHostname string

	// toggle skipping TLS on connections to github (for testing purposes)
	GithubSkipTLS bool
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/github-apps", h.get).Methods("GET")
	r.HandleFunc("/github-apps/new", h.new).Methods("GET")
	r.HandleFunc("/github-apps/exchange-code", h.exchangeCode).Methods("GET")
	r.HandleFunc("/github-apps/{github_app_id}/delete", h.delete).Methods("POST")
	r.HandleFunc("/github-apps/{github_app_id}/delete-install", h.deleteInstall).Methods("POST")
}

func (h *webHandlers) new(w http.ResponseWriter, r *http.Request) {
	type (
		hookAttrs struct {
			URL string `json:"url"`
		}
		manifest struct {
			Name        string            `json:"name"`
			URL         string            `json:"url"`
			Redirect    string            `json:"redirect_url"`
			SetupURL    string            `json:"setup_url"`
			Description string            `json:"description"`
			Events      []string          `json:"default_events"`
			Permissions map[string]string `json:"default_permissions"`
			Public      bool              `json:"public"`
			HookAttrs   hookAttrs         `json:"hook_attributes"`
		}
	)
	m := manifest{
		Name:        "otf-" + internal.GenerateRandomString(4),
		URL:         h.URL(""),
		SetupURL:    h.URL(paths.GithubApps()),
		HookAttrs:   hookAttrs{URL: h.URL(AppEventsPath)},
		Redirect:    h.URL(paths.ExchangeCodeGithubApp()),
		Description: "Trigger terraform runs in OTF from GitHub",
		Events:      []string{"push", "pull_request"},
		Public:      false,
		Permissions: map[string]string{
			"checks":        "write",
			"contents":      "read",
			"metadata":      "read",
			"pull_requests": "write",
			"statuses":      "write",
		},
	}
	marshaled, err := json.Marshal(&m)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("github_apps_new.tmpl", w, struct {
		html.SitePage
		Manifest       string
		GithubHostname string
	}{
		SitePage:       html.NewSitePage(r, "select app owner"),
		Manifest:       string(marshaled),
		GithubHostname: h.GithubHostname,
	})
}

func (h *webHandlers) get(w http.ResponseWriter, r *http.Request) {
	app, err := h.svc.GetGithubApp(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	installs, err := h.svc.ListInstallations(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := user.UserFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("github_apps_get.tmpl", w, struct {
		html.SitePage
		App            *App
		Installations  []*Installation
		GithubHostname string
		CanCreateApp   bool
		CanDeleteApp   bool
	}{
		SitePage:       html.NewSitePage(r, "github app"),
		App:            app,
		Installations:  installs,
		GithubHostname: h.GithubHostname,
		CanCreateApp:   user.CanAccessSite(rbac.CreateGithubAppAction),
		CanDeleteApp:   user.CanAccessSite(rbac.DeleteGithubAppAction),
	})
}

func (h *webHandlers) exchangeCode(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Code string `schema:"code,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// exchange code for credentials using an anonymous client
	client, err := NewClient(ClientOptions{
		Hostname:            h.GithubHostname,
		SkipTLSVerification: h.GithubSkipTLS,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	cfg, err := client.ExchangeCode(r.Context(), params.Code)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	opts := CreateAppOptions{
		AppID:         cfg.GetID(),
		Slug:          cfg.GetSlug(),
		WebhookSecret: cfg.GetWebhookSecret(),
		PrivateKey:    cfg.GetPEM(),
	}
	if cfg.GetOwner().GetType() == "Organization" {
		opts.Organization = cfg.GetOwner().Login
	}
	_, err = h.svc.CreateGithubApp(r.Context(), opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created github app: "+cfg.GetSlug())
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	app, err := h.svc.GetGithubApp(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.svc.DeleteGithubApp(r.Context()); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// render a small templated flash message
	buf := new(bytes.Buffer)
	err = h.RenderTemplate("github_delete_message.tmpl", buf, struct {
		GithubHostname string
		*App
	}{
		GithubHostname: h.GithubHostname,
		App:            app,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
	}
	html.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}

func (h *webHandlers) deleteInstall(w http.ResponseWriter, r *http.Request) {
	var params struct {
		InstallID int64 `schema:"install_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err := h.svc.DeleteInstallation(r.Context(), params.InstallID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, fmt.Sprintf("deleted installation: %d", params.InstallID))
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}
