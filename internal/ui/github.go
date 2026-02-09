package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
)

func addGithubAppHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/github-apps", h.getGithubApp).Methods("GET")
	r.HandleFunc("/github-apps/new", h.newGithubApp).Methods("GET")
	r.HandleFunc("/github-apps/exchange-code", h.exchangeCodeGithubApp).Methods("GET")
	r.HandleFunc("/github-apps/{github_app_id}/delete", h.deleteGithubApp).Methods("POST")
	r.HandleFunc("/github-apps/{github_app_id}/delete-install", h.deleteGithubAppInstall).Methods("POST")
}

func (h *Handlers) newGithubApp(w http.ResponseWriter, r *http.Request) {
	// Github manifest documented here:
	// https://docs.github.com/en/apps/sharing-github-apps/registering-a-github-app-from-a-manifest#implementing-the-github-app-manifest-flow
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
		URL:         h.HostnameService.URL(""),
		SetupURL:    h.HostnameService.URL(paths.GithubApps()),
		HookAttrs:   hookAttrs{URL: h.HostnameService.WebhookURL(github.AppEventsPath)},
		Redirect:    h.HostnameService.URL(paths.ExchangeCodeGithubApp()),
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
		html.Error(r, w, err.Error())
		return
	}

	props := newAppViewProps{
		manifest:       string(marshaled),
		githubHostname: h.GithubHostname.Host,
	}
	h.renderPage(
		h.templates.newAppView(props),
		"select app owner",
		w,
		r,
		withBreadcrumbs(helpers.Breadcrumb{Name: "Create GitHub app"}),
	)
}

func (h *Handlers) getGithubApp(w http.ResponseWriter, r *http.Request) {
	var installs []vcs.Installation
	app, err := h.GithubApp.GetApp(r.Context())
	if errors.Is(err, internal.ErrResourceNotFound) {
		// app not found, which is ok.
	} else if err != nil {
		html.Error(r, w, err.Error())
		return
	} else {
		// App found, now get installs
		installs, err = h.GithubApp.ListInstallations(r.Context())
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
	}

	props := getAppsProps{
		app:            app,
		installations:  installs,
		githubHostname: h.GithubHostname.Host,
		canCreateApp:   h.Authorizer.CanAccess(r.Context(), authz.CreateGithubAppAction, resource.SiteID),
		canDeleteApp:   h.Authorizer.CanAccess(r.Context(), authz.DeleteGithubAppAction, resource.SiteID),
	}
	h.renderPage(
		h.templates.getApps(props),
		"github app",
		w,
		r,
		withBreadcrumbs(helpers.Breadcrumb{Name: "GitHub app"}),
	)
}

func (h *Handlers) exchangeCodeGithubApp(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Code string `schema:"code,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// exchange code for credentials using an anonymous client
	client, err := github.NewClient(github.ClientOptions{
		BaseURL:             h.GithubHostname,
		SkipTLSVerification: h.SkipTLSVerification,
	})
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	cfg, err := client.ExchangeCode(r.Context(), params.Code)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	opts := github.CreateAppOptions{
		AppID:         cfg.GetID(),
		Slug:          cfg.GetSlug(),
		WebhookSecret: cfg.GetWebhookSecret(),
		PrivateKey:    cfg.GetPEM(),
		BaseURL:       h.GithubHostname,
	}
	if cfg.GetOwner().GetType() == "Organization" {
		opts.Organization = cfg.GetOwner().Login
	}
	_, err = h.GithubApp.CreateApp(r.Context(), opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "created github app: "+cfg.GetSlug())
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}

func (h *Handlers) deleteGithubApp(w http.ResponseWriter, r *http.Request) {
	app, err := h.GithubApp.GetApp(r.Context())
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	if err := h.GithubApp.DeleteApp(r.Context()); err != nil {
		html.Error(r, w, err.Error())
		return
	}
	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := deleteMessage(app).Render(r.Context(), buf); err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}

func (h *Handlers) deleteGithubAppInstall(w http.ResponseWriter, r *http.Request) {
	var params struct {
		InstallID int64 `schema:"install_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	err := h.GithubApp.DeleteInstallation(r.Context(), params.InstallID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, fmt.Sprintf("deleted installation: %d", params.InstallID))
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}
