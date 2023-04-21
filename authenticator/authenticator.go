// Package authenticator is responsible for handling the authentication of users with
// third party identity providers.
package authenticator

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
	"github.com/leg100/otf/tokens"
)

type (
	// authenticator logs people onto the system using an OAuth handshake with an
	// Identity provider before synchronising their user account and various organization
	// and team memberships from the provider.
	authenticator struct {
		otf.HostnameService
		tokens.TokensService // for creating session

		oauthClient
	}

	service struct {
		renderer       html.Renderer
		authenticators []*authenticator
	}

	Options struct {
		logr.Logger
		html.Renderer

		otf.HostnameService
		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		auth.AuthService
		tokens.TokensService

		Configs []cloud.CloudOAuthConfig
	}
)

func NewAuthenticatorService(opts Options) (*service, error) {
	svc := service{
		renderer: opts.Renderer,
	}

	for _, cfg := range opts.Configs {
		if cfg.OAuthConfig.ClientID == "" && cfg.OAuthConfig.ClientSecret == "" {
			// skip creating oauth client when creds are unspecified
			continue
		}
		client, err := NewOAuthClient(OAuthClientConfig{
			CloudOAuthConfig: cfg,
			otfHostname:      opts.HostnameService,
		})
		if err != nil {
			return nil, err
		}
		authenticator := &authenticator{
			HostnameService: opts.HostnameService,
			TokensService:   opts.TokensService,
			oauthClient:     client,
		}
		svc.authenticators = append(svc.authenticators, authenticator)

		opts.V(2).Info("activated oauth client", "name", cfg, "hostname", cfg.Hostname)
	}

	return &svc, nil
}

func (a *service) AddHandlers(r *mux.Router) {
	for _, authenticator := range a.authenticators {
		r.HandleFunc(authenticator.RequestPath(), authenticator.RequestHandler)
		r.HandleFunc(authenticator.CallbackPath(), authenticator.responseHandler)
	}
	r.HandleFunc("/login", a.loginHandler)
}

func (a *service) loginHandler(w http.ResponseWriter, r *http.Request) {
	a.renderer.Render("login.tmpl", w, struct {
		html.SitePage
		Authenticators []*authenticator
	}{
		SitePage:       html.NewSitePage(r, "login"),
		Authenticators: a.authenticators,
	})
}

// exchanging its auth code for a token.
func (a *authenticator) responseHandler(w http.ResponseWriter, r *http.Request) {
	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	token, err := a.CallbackHandler(r)
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Login(), http.StatusFound)
		return
	}

	client, err := a.NewClient(r.Context(), token)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// give authenticator unlimited access to services
	ctx := otf.AddSubjectToContext(r.Context(), &otf.Superuser{Username: "authenticator"})

	// Get cloud user
	cuser, err := client.GetUser(ctx)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = a.StartSession(w, r, tokens.StartSessionOptions{
		Username: &cuser.Name,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
