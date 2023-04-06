package tokens

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// session cookie stores the session token
const sessionCookie = "session"

type (
	// Aliases to disambiguate service names when embedded together.
	OrganizationService organization.Service

	TokensService interface {
		AgentTokenService
		tokenService

		StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error
		CreateRegistryToken(ctx context.Context, opts CreateRegistryTokenOptions) ([]byte, error)

		Middleware() mux.MiddlewareFunc
	}

	AgentTokenService interface {
		CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) ([]byte, error)
		GetAgentToken(ctx context.Context, id string) (*AgentToken, error)
		ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error)
		DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error)
	}

	tokenService interface {
		// CreateToken creates a user token.
		CreateToken(ctx context.Context, opts CreateTokenOptions) (*Token, []byte, error)
		// ListTokens lists API tokens for a user
		ListTokens(ctx context.Context) ([]*Token, error)
		// DeleteToken deletes a user token.
		DeleteToken(ctx context.Context, tokenID string) error
	}

	StartSessionOptions struct {
		Username *string
		Expiry   *time.Time
	}

	service struct {
		logr.Logger

		site         otf.Authorizer // authorizes site access
		organization otf.Authorizer // authorizes org access

		api *api
		db  *pgdb
		web *webHandlers

		middleware mux.MiddlewareFunc

		key jwk.Key
	}

	Options struct {
		logr.Logger
		otf.DB
		otf.Renderer
		auth.AuthService
		GoogleIAPConfig

		SiteToken string
		Secret    string
	}
)

func NewService(opts Options) (*service, error) {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &otf.SiteAuthorizer{Logger: opts.Logger},
		db:           &pgdb{opts.DB},
	}
	svc.api = &api{svc: &svc}
	svc.web = &webHandlers{
		Renderer:  opts.Renderer,
		svc:       &svc,
		siteToken: opts.SiteToken,
	}
	key, err := jwk.FromRaw([]byte(opts.Secret))
	if err != nil {
		return nil, err
	}
	svc.key = key
	svc.middleware = newMiddleware(middlewareOptions{
		AgentTokenService: &svc,
		AuthService:       opts.AuthService,
		GoogleIAPConfig:   opts.GoogleIAPConfig,
		SiteToken:         opts.SiteToken,
		key:               key,
	})

	return &svc, nil
}

func (a *service) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
	a.web.addHandlers(r)
}

// Middleware returns middleware for authenticating tokens
func (a *service) Middleware() mux.MiddlewareFunc { return a.middleware }

//
// Registry tokens service endpoints
//

func (a *service) CreateRegistryToken(ctx context.Context, opts CreateRegistryTokenOptions) ([]byte, error) {
	if opts.Organization == nil {
		return nil, fmt.Errorf("missing organization")
	}

	subject, err := a.organization.CanAccess(ctx, rbac.CreateRegistrySessionAction, *opts.Organization)
	if err != nil {
		return nil, err
	}

	expiry := otf.CurrentTimestamp().Add(defaultRegistrySessionExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	token, err := jwt.NewBuilder().
		Claim("kind", registrySessionKind).
		Claim("organization", *opts.Organization).
		IssuedAt(time.Now()).
		Expiration(expiry).
		Build()
	if err != nil {
		return nil, err
	}
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, a.key))
	if err != nil {
		return nil, err
	}

	a.V(2).Info("created registry session", "subject", subject, "run")

	return serialized, nil
}

//
// Agent tokens service endpoints
//

func (a *service) GetAgentToken(ctx context.Context, tokenID string) (*AgentToken, error) {
	at, err := a.db.GetAgentTokenByID(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}
	a.V(2).Info("retrieved agent token", "organization", at.Organization, "id", at.ID)
	return at, nil
}

func (a *service) CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) ([]byte, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateAgentTokenAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	at, token, err := NewAgentToken(NewAgentTokenOptions{
		CreateAgentTokenOptions: opts,
		key:                     a.key,
	})
	if err != nil {
		return nil, err
	}
	if err := a.db.CreateAgentToken(ctx, at); err != nil {
		a.Error(err, "creating agent token", "organization", opts.Organization, "id", at.ID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created agent token", "organization", opts.Organization, "id", at.ID, "subject", subject)
	return token, nil
}

func (a *service) ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListAgentTokensAction, organization)
	if err != nil {
		return nil, err
	}

	tokens, err := a.db.listAgentTokens(ctx, organization)
	if err != nil {
		a.Error(err, "listing agent tokens", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed agent tokens", "organization", organization, "subject", subject)
	return tokens, nil
}

func (a *service) DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error) {
	// retrieve agent token first in order to get organization for authorization
	at, err := a.db.GetAgentTokenByID(ctx, id)
	if err != nil {
		// we can't reveal any info because all we have is the
		// authentication token which is sensitive.
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteAgentTokenAction, at.Organization)
	if err != nil {
		return nil, err
	}

	if err := a.db.deleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "agent token", at, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted agent token", "agent token", at, "subject", subject)
	return at, nil
}

//
// User tokens service endpoints
//

// CreateToken creates a user token. Only users can create a user token, and
// they can only create a token for themselves.
func (a *service) CreateToken(ctx context.Context, opts CreateTokenOptions) (*Token, []byte, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	ut, token, err := NewToken(NewTokenOptions{
		CreateTokenOptions: opts,
		Username:           user.Username,
		key:                a.key,
	})
	if err != nil {
		a.Error(err, "constructing token", "user", user)
		return nil, nil, err
	}

	if err := a.db.CreateToken(ctx, ut); err != nil {
		a.Error(err, "creating token", "user", user)
		return nil, nil, err
	}

	a.V(1).Info("created token", "user", user)

	return ut, token, nil
}

func (a *service) ListTokens(ctx context.Context) ([]*Token, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return a.db.ListTokens(ctx, user.Username)
}

func (a *service) DeleteToken(ctx context.Context, tokenID string) error {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return err
	}

	token, err := a.db.GetToken(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving token", "user", user)
		return err
	}

	if user.Username != token.Username {
		return otf.ErrAccessNotPermitted
	}

	if err := a.db.DeleteToken(ctx, tokenID); err != nil {
		a.Error(err, "deleting token", "user", user)
		return err
	}

	a.V(1).Info("deleted token", "username", user)

	return nil
}

//
// User session service endpoints
//

func (a *service) StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error {
	if opts.Username == nil {
		return fmt.Errorf("missing username")
	}
	expiry := otf.CurrentTimestamp().Add(defaultExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	token, err := jwt.NewBuilder().
		Subject(*opts.Username).
		Claim("kind", userSessionKind).
		IssuedAt(time.Now()).
		Expiration(expiry).
		Build()
	if err != nil {
		return err
	}
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, a.key))
	if err != nil {
		return err
	}

	// Set cookie to expire at same time as token
	html.SetCookie(w, sessionCookie, string(serialized), otf.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	a.V(2).Info("started session", "username", *opts.Username)

	return nil
}
