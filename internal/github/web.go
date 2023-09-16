package github

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
)

type webHandlers struct {
	html.Renderer
	internal.HostnameService
	CloudService cloud.Service

	svc Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/admin/github-app", h.get).Methods("GET")
	r.HandleFunc("/admin/github-app/new", h.new).Methods("GET")
	r.HandleFunc("/admin/github-app/exchange-code", h.exchangeCode).Methods("POST")
	r.HandleFunc("/admin/github-app/complete", h.completeGithubSetup).Methods("POST")
	r.HandleFunc("/admin/github-app/delete", h.delete).Methods("POST")
}

func (h *webHandlers) new(w http.ResponseWriter, r *http.Request) {
	type (
		hookAttributes struct {
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
			HookAttrs   hookAttributes    `json:"hook_attributes"`
		}
	)
	m := manifest{
		Name: "OTF",
		URL:  fmt.Sprintf("https://%s", h.Hostname()),
		HookAttrs: hookAttributes{
			URL: fmt.Sprintf("https://%s/ghapp/webhook", h.Hostname()),
		},
		Redirect: fmt.Sprintf("https://%s%s", h.Hostname(),
			paths.ExchangeCodeGithubApp()),
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

	h.Render("github_list_apps.tmpl", w, struct {
		html.SitePage
		App *App
	}{
		SitePage: html.NewSitePage(r, "github app"),
		App:      app,
	})
}

func (h *webHandlers) exchangeCode(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string `schema:"organization_name,required"`
		Code         string `schema:"code,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// exchange code for credentials using an anonymous client
	client, err := NewClient(r.Context(), cloud.ClientOptions{})
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	cfg, err := client.ExchangeCode(r.Context(), params.Code)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err = h.svc.CreateGithubApp(r.Context(), CreateAppOptions{
		Organization:  params.Organization,
		AppID:         cfg.GetID(),
		WebhookSecret: cfg.GetWebhookSecret(),
		PrivateKey:    cfg.GetPEM(),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created github app: "+cfg.GetSlug())
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}

func (h *webHandlers) completeGithubSetup(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string `schema:"organization_name,required"`
		Code         string `schema:"code,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// exchange code for credentials using an anonymous client
	client, err := NewClient(r.Context(), cloud.ClientOptions{})
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	cfg, err := client.ExchangeCode(r.Context(), params.Code)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err = h.svc.CreateGithubApp(r.Context(), CreateAppOptions{
		Organization:  params.Organization,
		AppID:         cfg.GetID(),
		WebhookSecret: cfg.GetWebhookSecret(),
		PrivateKey:    cfg.GetPEM(),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created github app: "+cfg.GetSlug())
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteGithubApp(r.Context()); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted github app")
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}
