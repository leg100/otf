package registry

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type service interface {
	create(ctx context.Context, organization string) (*Session, error)
	get(ctx context.Context, token string) (*Session, error)
}

// app is the implementation of service
type app struct {
	otf.Authorizer
	logr.Logger

	db
	*handlers
}

func NewApplication(ctx context.Context, opts ApplicationOptions) *app {
	app := &app{
		Authorizer: opts.Authorizer,
		Logger:     opts.Logger,
		db:         newDB(ctx, opts.Database, defaultExpiry),
	}
	app.handlers = &handlers{
		app: app,
	}
	return app
}

type ApplicationOptions struct {
	otf.Authorizer
	otf.Database
	logr.Logger
}

// CreateRegistrySession creates and persists a registry session.
func (a *app) CreateRegistrySession(ctx context.Context, organization string) (otf.RegistrySession, error) {
	return a.create(ctx, organization)
}

// GetRegistrySession retrieves a registry session using a token. Useful for
// checking token is valid.
func (a *app) GetRegistrySession(ctx context.Context, token string) (*Session, error) {
	return a.get(ctx, token)
}

func (a *app) create(ctx context.Context, organization string) (*Session, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateRegistrySessionAction, organization)
	if err != nil {
		return nil, err
	}

	session, err := newSession(organization)
	if err != nil {
		a.Error(err, "constructing registry session", "subject", subject, "organization", organization)
		return nil, err
	}
	if err := a.db.create(ctx, session); err != nil {
		a.Error(err, "creating registry session", "subject", subject, "session", session)
		return nil, err
	}

	a.V(2).Info("created registry session", "subject", subject, "session", session)

	return session, nil
}

func (a *app) get(ctx context.Context, token string) (*Session, error) {
	// No need for authz because caller is providing an auth token.

	session, err := a.db.get(ctx, token)
	if err != nil {
		a.Error(err, "retrieving registry session", "token", "*****")
		return nil, err
	}

	a.V(2).Info("retrieved registry session", "session", session)

	return session, nil
}
