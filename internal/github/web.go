package github

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
)

const githubAppConfigCookie = "github-app-config"

type webHandlers struct {
	html.Renderer
	internal.HostnameService
	CloudService

	svc Service
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/github-apps", h.listGithubApps)
	r.HandleFunc("/organizations/{organization_name}/github-apps/new", h.newGithubApp)
	r.HandleFunc("/organizations/{organization_name}/github-apps/exchange-code", h.githubAppExchangeCode)
	r.HandleFunc("/organizations/{organization_name}/github-apps/complete", h.completeGithubSetup)
	r.HandleFunc("/github-apps/{github_app_id}/delete", h.deleteGithubApp)

	r.HandleFunc("/github-apps/{github_app_id}/new-install", h.newGithubAppInstall)
	r.HandleFunc("/github-apps/{github_app_id}/create-install", h.createGithubAppInstall)
}

func (h *webHandlers) new(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string `schema:"organization_name,required"`
		Cloud        string `schema:"cloud,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tmpl := fmt.Sprintf("vcs_provider_%s_new.tmpl", params.Cloud)
	h.Render(tmpl, w, struct {
		organization.OrganizationPage
		Cloud string
	}{
		OrganizationPage: organization.NewPage(r, "new vcs provider", params.Organization),
		Cloud:            params.Cloud,
	})
}

func (h *webHandlers) create(w http.ResponseWriter, r *http.Request) {
	var params struct {
		OrganizationName string `schema:"organization_name,required"`
		Token            string `schema:"token,required"`
		Name             string `schema:"name,required"`
		Cloud            string `schema:"cloud,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.CreateVCSProvider(r.Context(), CreateOptions{
		Organization: params.OrganizationName,
		Token:        &params.Token,
		Name:         params.Name,
		Cloud:        params.Cloud,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created provider: "+provider.Name)
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	providers, err := h.svc.ListVCSProviders(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("vcs_provider_list.tmpl", w, struct {
		organization.OrganizationPage
		Items        []*VCSProvider
		CloudConfigs []cloud.Config
	}{
		OrganizationPage: organization.NewPage(r, "vcs providers", org),
		Items:            providers,
		CloudConfigs:     h.ListCloudConfigs(),
	})
}

func (h *webHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("vcs_provider_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.GetVCSProvider(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("vcs_provider_get.tmpl", w, struct {
		VCSProvider *VCSProvider
	}{
		VCSProvider: provider,
	})
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("vcs_provider_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.DeleteVCSProvider(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted provider: "+provider.Name)
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *webHandlers) newGithubApp(w http.ResponseWriter, r *http.Request) {
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

func (h *webHandlers) listGithubApps(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	providers, err := h.svc.ListVCSProviders(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("vcs_provider_list.tmpl", w, struct {
		organization.OrganizationPage
		Items        []*VCSProvider
		CloudConfigs []cloud.Config
	}{
		OrganizationPage: organization.NewPage(r, "vcs providers", org),
		Items:            providers,
		CloudConfigs:     h.ListCloudConfigs(),
	})
}

func (h *webHandlers) githubAppExchangeCode(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string `schema:"organization_name,required"`
		Code         string `schema:"code,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// exchange code for credentials using an anonymous client
	client, err := github.NewClient(r.Context(), cloud.ClientOptions{})
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	cfg, err := client.ExchangeCode(r.Context(), params.Code)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.CreateVCSProvider(r.Context(), CreateOptions{
		Organization: params.Organization,
		Name:         cfg.GetSlug(),
		Cloud:        cloud.Github,
		GithubApp: newGithubApp(newGithubAppOptions{
			AppID:         cfg.GetID(),
			WebhookSecret: cfg.GetWebhookSecret(),
			PrivateKey:    cfg.GetPEM(),
		}),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created github app: "+cfg.GetSlug())
	http.Redirect(w, r, paths.VCSProvider(provider.ID), http.StatusFound)
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
	client, err := github.NewClient(r.Context(), cloud.ClientOptions{})
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	cfg, err := client.ExchangeCode(r.Context(), params.Code)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.CreateVCSProvider(r.Context(), CreateOptions{
		Organization: params.Organization,
		Name:         cfg.GetSlug(),
		Cloud:        cloud.Github,
		GithubApp: newGithubApp(newGithubAppOptions{
			AppID:         cfg.GetID(),
			WebhookSecret: cfg.GetWebhookSecret(),
			PrivateKey:    cfg.GetPEM(),
		}),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created github app: "+cfg.GetSlug())
	http.Redirect(w, r, paths.VCSProvider(provider.ID), http.StatusFound)
}
