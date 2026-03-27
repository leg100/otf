package ui

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/vcs"
)

type Handlers struct {
	GithubApp           GithubAppService
	Hostnames           HostnameService
	GithubHostname      *internal.WebURL
	SkipTLSVerification bool
	Authorizer          authz.Interface
	templates           *templates
}

type GithubAppService interface {
	CreateApp(context.Context, github.CreateAppOptions) (*github.App, error)
	GetApp(context.Context) (*github.App, error)
	DeleteApp(context.Context) error
	ListInstallations(context.Context) ([]vcs.Installation, error)
	DeleteInstallation(context.Context, int64) error
}

type HostnameService interface {
	URL(path string) string
	WebhookURL(path string) string
}

type templates struct{}

func NewHandlers(
	githubApp GithubAppService,
	hostnames HostnameService,
	githubHostname *internal.WebURL,
	skipTLSVerification bool,
	authorizer authz.Interface,
) *Handlers {
	return &Handlers{
		GithubApp:           githubApp,
		Hostnames:           hostnames,
		GithubHostname:      githubHostname,
		SkipTLSVerification: skipTLSVerification,
		Authorizer:          authorizer,
		templates:           &templates{},
	}
}

func (h *Handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/github-apps", h.getGithubApp).Methods("GET")
	r.HandleFunc("/github-apps/new", h.newGithubApp).Methods("GET")
	r.HandleFunc("/github-apps/exchange-code", h.exchangeCodeGithubApp).Methods("GET")
	r.HandleFunc("/github-apps/{github_app_id}/delete", h.deleteGithubApp).Methods("POST")
	r.HandleFunc("/github-apps/{github_app_id}/delete-install", h.deleteGithubAppInstall).Methods("POST")
}
