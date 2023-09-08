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
	"github.com/leg100/otf/internal/organization"
)

type webHandlers struct {
	html.Renderer
	internal.HostnameService
	CloudService cloud.Service

	svc Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/github-apps", h.list).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/github-apps/new", h.new).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/github-apps/exchange-code", h.exchangeCode).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/github-apps/complete", h.completeGithubSetup).Methods("POST")
	r.HandleFunc("/github-apps/{github_app_id}/delete", h.delete).Methods("POST")

	// prompt user to add github app installation to OTF as a VCS provider
	r.HandleFunc("/github-apps/{github_app_id}/new-install", h.newInstall).Methods("GET")
	r.HandleFunc("/github-apps/{github_app_id}/create-install", h.createInstall).Methods("POST")
}

func (h *webHandlers) new(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

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
			paths.ExchangeCodeGithubApp(org)),
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

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	apps, err := h.svc.ListGithubApps(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("github_list_apps.tmpl", w, struct {
		organization.OrganizationPage
		Items []*GithubApp
	}{
		OrganizationPage: organization.NewPage(r, "github apps", org),
		Items:            apps,
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

	app, err := h.svc.CreateGithubApp(r.Context(), CreateAppOptions{
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
	http.Redirect(w, r, paths.GithubApp(app.ID), http.StatusFound)
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

	app, err := h.svc.CreateGithubApp(r.Context(), CreateAppOptions{
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
	http.Redirect(w, r, paths.GithubApp(app.ID), http.StatusFound)
}

func (h *webHandlers) newInstall(w http.ResponseWriter, r *http.Request) {
	var params struct {
		AppID     string `schema:"github_app_id,required"`
		InstallID string `schema:"installation_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	app, err := h.svc.DeleteGithubApp(r.Context(), params.AppID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted app: "+app.ID)
	http.Redirect(w, r, paths.GithubApps(app.Organization), http.StatusFound)
}

func (h *webHandlers) createInstall(w http.ResponseWriter, r *http.Request) {
	var params struct {
		AppID     string `schema:"github_app_id,required"`
		InstallID string `schema:"installation_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	app, err := h.svc.DeleteGithubApp(r.Context(), params.AppID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted app: "+app.ID)
	http.Redirect(w, r, paths.GithubApps(app.Organization), http.StatusFound)
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("github_app_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	app, err := h.svc.DeleteGithubApp(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted app: "+app.ID)
	http.Redirect(w, r, paths.GithubApps(app.Organization), http.StatusFound)
}
