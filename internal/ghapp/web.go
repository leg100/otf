package ghapp

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
)

// URL path for the endpoint receiving vcs events
const webhookPath = "/webhooks/ghapp"

type (
	WebHandlers struct {
		html.Renderer
		cloud.Service
		internal.HostnameService
	}
)

func (h *WebHandlers) AddHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/ghapps/new", h.new)
	r.HandleFunc("/organizations/{organization_name}/ghapps/exchange-code", h.exchangeCode)
	r.HandleFunc("/organizations/{organization_name}/ghapps", h.list)
}

func (h *WebHandlers) new(w http.ResponseWriter, r *http.Request) {
	type (
		hookAttributes struct {
			URL string `json:"url"`
		}
		manifest struct {
			Name        string            `json:"name"`
			URL         string            `json:"url"`
			Redirect    string            `json:"redirect_url"`
			Description string            `json:"description"`
			Events      []string          `json:"default_events"`
			Permissions map[string]string `json:"default_permissions"`
			Public      bool              `json:"public"`
			HookAttrs   hookAttributes    `json:"hook_attributes"`
		}
	)
	m := manifest{
		Name: "OTF",
		URL:  (&url.URL{Scheme: "https", Host: h.Hostname()}).String(),
		HookAttrs: hookAttributes{
			URL: (&url.URL{
				Scheme: "https",
				Host:   h.Hostname(),
				Path:   webhookPath,
			}).String(),
		},
		Redirect: (&url.URL{
			Scheme: "https",
			Host:   h.Hostname(),
			Path:   paths.ExchangeCodeGithubApp(),
		}).String(),
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

func (h *WebHandlers) exchangeCode(w http.ResponseWriter, r *http.Request) {
	code, err := decode.Param("code", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// exchange code for credentials
	client, err := github.NewClient(r.Context(), cloud.ClientOptions{})
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	cfg, err := client.ExchangeCode(r.Context(), code)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	marshalled, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	w.Write(marshalled)
}

func (h *WebHandlers) list(w http.ResponseWriter, r *http.Request) {}
