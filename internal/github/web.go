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
)

const (
	// GithubPath is the URL path for the endpoint receiving VCS events from the
	// Github App
	AppEventsPath = "/webhooks/vcs/github-app"
)

type webHandlers struct {
	html.Renderer
	internal.HostnameService

	svc Service
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
		HookAttrs:   hookAttrs{URL: h.URL(AppEventsPath)},
		Redirect:    h.URL(paths.ExchangeCodeGithubApp()),
		Description: "Trigger terraform runs in OTF from GitHub",
		Events:      []string{"push", "pull_request"},
		Public:      false,
		Permissions: map[string]string{
			"checks":        "write",
			"contents":      "read",
			"pull_requests": "write",
			"statuses":      "write",
		},
	}
	marshaled, err := json.Marshal(&m)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("ghapp_register.tmpl", w, struct {
		html.SitePage
		Manifest string
	}{
		SitePage: html.NewSitePage(r, "select app owner"),
		Manifest: string(marshaled),
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

	h.Render("github_apps_get.tmpl", w, struct {
		html.SitePage
		App           *App
		Installations []*Installation
	}{
		SitePage:      html.NewSitePage(r, "github app"),
		App:           app,
		Installations: installs,
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
	client, err := NewClient(r.Context(), ClientOptions{})
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
	err = h.RenderTemplate("github_delete_message.tmpl", buf, app)
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
