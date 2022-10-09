package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateUser(ctx context.Context, username string) (*otf.User, error) {
	user := otf.NewUser(username)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username)
		return nil, err
	}

	a.V(1).Info("created user", "username", username)

	return user, nil
}

// EnsureCreatedUser retrieves the user or creates the user if they don't exist.
func (a *Application) EnsureCreatedUser(ctx context.Context, username string) (*otf.User, error) {
	user, err := a.db.GetUser(ctx, otf.UserSpec{Username: &username})
	if err == nil {
		return user, nil
	}
	if err != otf.ErrResourceNotFound {
		a.Error(err, "retrieving user", "username", username)
		return nil, err
	}

	return a.CreateUser(ctx, username)
}

func (a *Application) SyncUserMemberships(ctx context.Context, user *otf.User, orgs []*otf.Organization, teams []*otf.Team) (*otf.User, error) {
	if err := user.SyncMemberships(ctx, a.db, orgs, teams); err != nil {
		return nil, err
	}

	a.V(1).Info("synchronised user's memberships", "username", user.Username())

	return user, nil
}

// CreateSession creates a session and adds it to the user.
func (a *Application) CreateSession(ctx context.Context, user *otf.User, data *otf.SessionData) (*otf.Session, error) {
	session, err := user.AttachNewSession(data)
	if err != nil {
		a.Error(err, "attaching session", "username", user.Username())
		return nil, err
	}

	if err := a.db.CreateSession(ctx, session); err != nil {
		a.Error(err, "creating session", "username", user.Username())
		return nil, err
	}

	a.V(1).Info("created session", "username", user.Username())

	return session, nil
}

func (a *Application) GetUser(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	user, err := a.db.GetUser(ctx, spec)
	if err != nil {
		// Failure to retrieve a user is frequently due to the fact the http
		// middleware first calls this endpoint to see if the bearer token
		// belongs to a user and if that fails it then checks if it belongs to
		// an agent. Therefore we log this at a low priority info level rather
		// than as an error.
		//
		// TODO: make bearer token a signed cryptographic token containing
		// metadata about the authenticating entity.
		info := append([]any{"err", err.Error()}, spec.KeyValue()...)
		a.V(2).Info("unable to retrieve user", info...)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username())

	return user, nil
}

func (a *Application) DeleteSession(ctx context.Context, token string) error {
	// Retrieve user purely for logging purposes
	user, err := a.GetUser(ctx, otf.UserSpec{SessionToken: &token})
	if err != nil {
		return err
	}

	if err := a.db.DeleteSession(ctx, token); err != nil {
		a.Error(err, "deleting session", "username", user.Username())
		return err
	}

	a.V(1).Info("deleted session", "username", user.Username())

	return nil
}

// CreateToken creates a user token.
func (a *Application) CreateToken(ctx context.Context, user *otf.User, opts *otf.TokenCreateOptions) (*otf.Token, error) {
	token, err := otf.NewToken(user.ID(), opts.Description)
	if err != nil {
		a.Error(err, "constructing token", "username", user.Username())
		return nil, err
	}

	if err := a.db.CreateToken(ctx, token); err != nil {
		a.Error(err, "creating token", "username", user.Username())
		return nil, err
	}

	a.V(1).Info("created token", "username", user.Username())

	return token, nil
}

func (a *Application) DeleteToken(ctx context.Context, user *otf.User, tokenID string) error {
	if err := a.db.DeleteToken(ctx, tokenID); err != nil {
		a.Error(err, "deleting token", "username", user.Username())
		return err
	}

	a.V(1).Info("deleted token", "username", user.Username())

	return nil
}
