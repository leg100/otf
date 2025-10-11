package ui

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
)

type handlers struct {
	hostnameService     *internal.HostnameService
	authorizer          *authz.Authorizer
	skipTLSVerification bool

	*githubApp
}

type Options struct {
	HostnameService *internal.HostnameService
	Authorizer      *authz.Authorizer

	GithubAPIURL        *internal.WebURL
	GithubAppService    githubAppClient
	SkipTLSVerification bool
}

func New(opts Options) *handlers {
	h := &handlers{
		authorizer:          opts.Authorizer,
		skipTLSVerification: opts.SkipTLSVerification,
	}
	h.githubApp = &githubApp{
		handlers:     h,
		githubAPIURL: opts.GithubAPIURL,
		svc:          opts.GithubAppService,
	}
	return h
}

func (a *handlers) AddHandlers(r *mux.Router) {
	a.githubApp.addHandlers(r)
}
