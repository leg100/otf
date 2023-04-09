// Package authenticator is responsible for handling the authentication of users with
// third party identity providers.
package authenticator

import (
	"context"
	"github.com/coreos/go-oidc"
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
	"golang.org/x/oauth2"
	"net/http"
)

type (
	authenticator interface {
		RequestPath() string
		CallbackPath() string
		RequestHandler(w http.ResponseWriter, r *http.Request)
		ResponseHandler(w http.ResponseWriter, r *http.Request)
	}

	// oauthAuthenticator logs people onto the system using an OAuth handshake with an
	// Identity provider before synchronising their user account and various organization
	// and team memberships from the provider.
	oauthAuthenticator struct {
		otf.HostnameService
		tokens.TokensService // for creating session

		oauthClient
	}

	service struct {
		renderer       otf.Renderer
		authenticators []authenticator
	}

	Options struct {
		logr.Logger
		otf.Renderer

		otf.HostnameService
		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		auth.AuthService
		tokens.TokensService

		Configs     []cloud.CloudOAuthConfig
		OIDCConfigs []cloud.OIDCConfig
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
		authenticator := &oauthAuthenticator{
			HostnameService: opts.HostnameService,
			TokensService:   opts.TokensService,
			oauthClient:     client,
		}
		svc.authenticators = append(svc.authenticators, authenticator)

		opts.V(2).Info("activated oauth client", "name", cfg, "hostname", cfg.Hostname)
	}

	for _, cfg := range opts.OIDCConfigs {
		if cfg.ClientID == "" && cfg.ClientSecret == "" {
			// skip creating oidc client when creds are unspecified
			continue
		}

		provider, err := oidc.NewProvider(context.Background(), cfg.IssuerURL)
		if err != nil {
			return nil, err
		}

		verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

		authenticator := &oidcAuthenticator{
			TokensService: opts.TokensService,
			oidcConfig:    cfg,
			provider:      provider,
			verifier:      verifier,
			oauth2Config: oauth2.Config{
				ClientID:     cfg.ClientID,
				ClientSecret: cfg.ClientSecret,
				RedirectURL:  cfg.RedirectURL,
				Endpoint:     provider.Endpoint(),

				// "openid" is a required scope for OpenID Connect flows.
				// groups is used for managing permissions.
				Scopes: append(cfg.Scopes, oidc.ScopeOpenID, "groups", "profile"),
			},
		}
		svc.authenticators = append(svc.authenticators, authenticator)

		opts.V(2).Info("activated oidc client", "name", cfg, "issuerURL", cfg.IssuerURL, "redirectURL", cfg.RedirectURL, "name", cfg.Name)
	}

	return &svc, nil
}

func (a *service) AddHandlers(r *mux.Router) {
	for _, authenticator := range a.authenticators {
		r.HandleFunc(authenticator.RequestPath(), authenticator.RequestHandler)
		r.HandleFunc(authenticator.CallbackPath(), authenticator.ResponseHandler)
	}
	r.HandleFunc("/login", a.loginHandler)
}

func (a *service) loginHandler(w http.ResponseWriter, r *http.Request) {
	a.renderer.Render("login.tmpl", w, r, a.authenticators)
}

// ResponseHandler handles exchanging its auth code for a token.
func (a *oauthAuthenticator) ResponseHandler(w http.ResponseWriter, r *http.Request) {
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

	// give oauthAuthenticator unlimited access to services
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
