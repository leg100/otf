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

		HostnameService      *internal.HostnameService
		UserService          userService
		TokensService        *tokens.Service
		IDTokenHandlerConfig OIDCConfig
		SkipTLSVerification  bool
	}

	Service struct {
		logr.Logger

		HostnameService *internal.HostnameService
		UserService     userService
		TokensService   *tokens.Service
		clients         []*OAuthClient
	}

	userService interface {
		GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error)
		Create(ctx context.Context, username string, opts ...user.NewUserOption) (*user.User, error)
		UpdateAvatar(ctx context.Context, username user.Username, avatarURL string) error
	}
)

// NewAuthenticatorService constructs a service for logging users onto
// the system. Supports multiple clients: zero or more clients that support an
// opaque token, and one client that supports IDToken/OIDC.
func NewAuthenticatorService(ctx context.Context, opts Options) (*Service, error) {
	svc := Service{
		Logger:          opts.Logger,
		HostnameService: opts.HostnameService,
		UserService:     opts.UserService,
		TokensService:   opts.TokensService,
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
		opts.Logger,
		handler,
		opts.HostnameService,
		opts.TokensService,
		opts.UserService,
		OAuthConfig{
			Endpoint:            handler.provider.Endpoint(),
			Scopes:              opts.IDTokenHandlerConfig.Scopes,
			ClientID:            opts.IDTokenHandlerConfig.ClientID,
			ClientSecret:        opts.IDTokenHandlerConfig.ClientSecret,
			Name:                opts.IDTokenHandlerConfig.Name,
			SkipTLSVerification: opts.SkipTLSVerification,
			Icon:                oidcIcon(),
		},
	)
	if err != nil {
		return nil, err
	}
	svc.clients = append(svc.clients, client)
	opts.V(0).Info("activated OIDC client", "name", opts.IDTokenHandlerConfig.Name)
	return &svc, nil
}

func (a *Service) AddHandlers(r *mux.Router) {
	for _, authenticator := range a.clients {
		authenticator.addHandlers(r)
	}
	r.HandleFunc("/login", a.loginHandler)
}

func (a *Service) RegisterOAuthClient(cfg OpaqueHandlerConfig) error {
	// Construct clients with opaque token handlers
	if cfg.ClientID == "" && cfg.ClientSecret == "" {
		// skip creating OAuth client when creds are unspecified
		return nil
	}
	client, err := newOAuthClient(
		a.Logger,
		&opaqueHandler{cfg},
		a.HostnameService,
		a.TokensService,
		a.UserService,
		cfg.OAuthConfig,
	)
	if err != nil {
		return err
	}
	a.clients = append(a.clients, client)
	a.V(0).Info("activated OAuth client", "name", cfg.Name, "hostname", cfg.BaseURL)

	return nil
}

func (a *Service) loginHandler(w http.ResponseWriter, r *http.Request) {
	html.Render(login(a.clients), w, r)
}
