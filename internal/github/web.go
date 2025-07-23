package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
)

const (
	// GithubPath is the URL path for the endpoint receiving VCS events from the
	// Github App
	AppEventsPath = "/webhooks/github-app"
)

type webHandlers struct {
	*internal.HostnameService

	svc        webClient
	authorizer *authz.Authorizer

	githubAPIURL        *internal.WebURL
	skipTLSVerification bool
}

// webClient provides web handlers with access to github app service endpoints
type webClient interface {
	CreateApp(ctx context.Context, opts CreateAppOptions) (*App, error)
	GetApp(ctx context.Context) (*App, error)
	DeleteApp(ctx context.Context) error

	ListInstallations(ctx context.Context) ([]vcs.Installation, error)
	DeleteInstallation(ctx context.Context, installID int64) error
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
		URL:         h.URL(""),
		SetupURL:    h.URL(paths.GithubApps()),
		HookAttrs:   hookAttrs{URL: h.WebhookURL(AppEventsPath)},
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
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := newAppViewProps{
		manifest:       string(marshaled),
		githubHostname: h.githubAPIURL.Host,
	}
	html.Render(newAppView(props), w, r)
}

func (h *webHandlers) get(w http.ResponseWriter, r *http.Request) {
	var installs []vcs.Installation
	app, err := h.svc.GetApp(r.Context())
	if errors.Is(err, internal.ErrResourceNotFound) {
		// app not found, which is ok.
	} else if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		// App found, now get installs
		installs, err = h.svc.ListInstallations(r.Context())
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	props := getAppsProps{
		app:            app,
		installations:  installs,
		githubHostname: h.githubAPIURL.Host,
		canCreateApp:   h.authorizer.CanAccess(r.Context(), authz.CreateGithubAppAction, resource.SiteID),
		canDeleteApp:   h.authorizer.CanAccess(r.Context(), authz.DeleteGithubAppAction, resource.SiteID),
	}
	html.Render(getApps(props), w, r)
}

func (h *webHandlers) exchangeCode(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Code string `schema:"code,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// exchange code for credentials using an anonymous client
	client, err := NewClient(ClientOptions{
		BaseURL:             h.githubAPIURL,
		SkipTLSVerification: h.skipTLSVerification,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	cfg, err := client.ExchangeCode(r.Context(), params.Code)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	opts := CreateAppOptions{
		AppID:         cfg.GetID(),
		Slug:          cfg.GetSlug(),
		WebhookSecret: cfg.GetWebhookSecret(),
		PrivateKey:    cfg.GetPEM(),
		BaseURL:       h.githubAPIURL,
	}
	if cfg.GetOwner().GetType() == "Organization" {
		opts.Organization = cfg.GetOwner().Login
	}
	_, err = h.svc.CreateApp(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created github app: "+cfg.GetSlug())
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	app, err := h.svc.GetApp(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.svc.DeleteApp(r.Context()); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := deleteMessage(app).Render(r.Context(), buf); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}

func (h *webHandlers) deleteInstall(w http.ResponseWriter, r *http.Request) {
	var params struct {
		InstallID int64 `schema:"install_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err := h.svc.DeleteInstallation(r.Context(), params.InstallID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, fmt.Sprintf("deleted installation: %d", params.InstallID))
	http.Redirect(w, r, paths.GithubApps(), http.StatusFound)
}
