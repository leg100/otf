package authenticator

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/http/html"
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

		authenticator, err := newOIDCAuthenticator(context.Background(), oidcAuthenticatorOptions{
			TokensService:   opts.TokensService,
			HostnameService: opts.HostnameService,
			OIDCConfig:      cfg,
		})
		if err != nil {
			return nil, err
		}

		svc.authenticators = append(svc.authenticators, authenticator)

		opts.V(0).Info("activated oidc client", "name", cfg.Name)
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
	a.renderer.Render("login.tmpl", w, struct {
		html.SitePage
		Authenticators []authenticator
	}{
		SitePage:       html.NewSitePage(r, "login"),
		Authenticators: a.authenticators,
	})
}
