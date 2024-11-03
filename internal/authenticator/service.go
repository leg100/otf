package authenticator

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/user"
)

type (
	Options struct {
		logr.Logger
		html.Renderer

		*internal.HostnameService

		UserService          userService
		TokensService        *tokens.Service
		OpaqueHandlerConfigs []OpaqueHandlerConfig
		IDTokenHandlerConfig OIDCConfig
		SkipTLSVerification  bool
	}

	service struct {
		html.Renderer

		clients []*OAuthClient
	}

	userService interface {
		GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error)
		Create(ctx context.Context, username string) (*user.User, error)
	}
)

// NewAuthenticatorService constructs a service for logging users onto
// the system. Supports multiple clients: zero or more clients that support an
// opaque token, and one client that supports IDToken/OIDC.
func NewAuthenticatorService(ctx context.Context, opts Options) (*service, error) {
	svc := service{Renderer: opts.Renderer}
	// Construct clients with opaque token handlers
	for _, cfg := range opts.OpaqueHandlerConfigs {
		if cfg.ClientID == "" && cfg.ClientSecret == "" {
			// skip creating OAuth client when creds are unspecified
			continue
		}
		cfg.SkipTLSVerification = opts.SkipTLSVerification
		client, err := newOAuthClient(
			&opaqueHandler{cfg},
			opts.HostnameService,
			opts.TokensService,
			cfg.OAuthConfig,
		)
		if err != nil {
			return nil, err
		}
		svc.clients = append(svc.clients, client)
		opts.V(0).Info("activated OAuth client", "name", cfg.Name, "hostname", cfg.Hostname)
	}
	// Construct client with OIDC IDToken handler
	if opts.IDTokenHandlerConfig.ClientID == "" && opts.IDTokenHandlerConfig.ClientSecret == "" {
		// skip creating OIDC authenticator when creds are unspecified
		return &svc, nil
	}
	opts.IDTokenHandlerConfig.SkipTLSVerification = opts.SkipTLSVerification
	handler, err := newIDTokenHandler(ctx, opts.IDTokenHandlerConfig)
	if err != nil {
		return nil, err
	}
	client, err := newOAuthClient(
		handler,
		opts.HostnameService,
		opts.TokensService,
		OAuthConfig{
			Endpoint:            handler.provider.Endpoint(),
			Scopes:              opts.IDTokenHandlerConfig.Scopes,
			ClientID:            opts.IDTokenHandlerConfig.ClientID,
			ClientSecret:        opts.IDTokenHandlerConfig.ClientSecret,
			Name:                opts.IDTokenHandlerConfig.Name,
			SkipTLSVerification: opts.SkipTLSVerification,
		},
	)
	if err != nil {
		return nil, err
	}
	svc.clients = append(svc.clients, client)
	opts.V(0).Info("activated OIDC client", "name", opts.IDTokenHandlerConfig.Name)
	return &svc, nil
}

func (a *service) AddHandlers(r *mux.Router) {
	for _, authenticator := range a.clients {
		authenticator.addHandlers(r)
	}
	r.HandleFunc("/login", a.loginHandler)
}

func (a *service) loginHandler(w http.ResponseWriter, r *http.Request) {
	a.Render("login.tmpl", w, struct {
		html.SitePage
		Clients []*OAuthClient
	}{
		SitePage: html.NewSitePage(r, "login"),
		Clients:  a.clients,
	})
}
