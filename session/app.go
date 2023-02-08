package session

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

type app interface {
	otf.SessionService

	get(ctx context.Context, token string) (*Session, error)
	list(ctx context.Context, userID string) ([]*Session, error)
	delete(ctx context.Context, token string) error
}

type Application struct {
	otf.Authorizer
	logr.Logger

	db db
	*handlers
	*htmlApp
}

func NewApplication(opts ApplicationOptions) *Application {
	db := newPGDB(opts.Database)

	// purge expired sessions
	go db.startSessionExpirer(defaultCleanupInterval)

	app := &Application{
		Authorizer: opts.Authorizer,
		db:         db,
		Logger:     opts.Logger,
	}
	app.handlers = &handlers{
		Application: app,
	}
	app.htmlApp = &htmlApp{
		app:      app,
		Renderer: opts.Renderer,
	}
	return app
}

type ApplicationOptions struct {
	otf.Authorizer
	otf.Database
	otf.Renderer
	logr.Logger
}

// CreateSession creates and persists a user session.
func (a *Application) CreateSession(r *http.Request, userID string) (otf.Session, error) {
	session, err := NewSession(r, userID)
	if err != nil {
		a.Error(err, "building new session", "uid", userID)
		return nil, err
	}
	if err := a.db.CreateSession(r.Context(), session); err != nil {
		a.Error(err, "creating session", "uid", userID)
		return nil, err
	}

	a.V(2).Info("created session", "uid", userID)

	return session, nil
}

func (a *Application) get(ctx context.Context, token string) (otf.Session, error) {
	return a.db.GetSessionByToken(ctx, token)
}

func (a *Application) list(ctx context.Context, userID string) ([]*Session, error) {
	return a.db.ListSessions(ctx, userID)
}

func (a *Application) delete(ctx context.Context, token string) error {
	if err := a.db.DeleteSession(ctx, token); err != nil {
		a.Error(err, "deleting session")
		return err
	}

	a.V(2).Info("deleted session")

	return nil
}
