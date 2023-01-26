package app

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

func (a *Application) CreateUser(ctx context.Context, username string) (*otf.User, error) {
	user := otf.NewUser(username)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username)
		return nil, err
	}

	a.V(0).Info("created user", "username", username)

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

// ListUsers lists users with various filters.
func (a *Application) ListUsers(ctx context.Context, opts otf.UserListOptions) ([]*otf.User, error) {
	var err error
	if opts.Organization != nil && opts.TeamName != nil {
		// team members can view membership of their own teams.
		subject, err := otf.SubjectFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if user, ok := subject.(*otf.User); ok && !user.IsSiteAdmin() {
			_, err := user.Team(*opts.TeamName, *opts.Organization)
			if err != nil {
				// user is not a member of the team
				return nil, otf.ErrAccessNotPermitted
			}
			return a.db.ListUsers(ctx, opts)
		}
	}

	if opts.Organization != nil {
		// subject needs perms on org to list users in org
		_, err = a.CanAccessOrganization(ctx, rbac.ListUsersAction, *opts.Organization)
	} else {
		// subject needs perms on site to list users across site
		_, err = a.CanAccessSite(ctx, rbac.ListRunsAction)
	}
	if err != nil {
		return nil, err
	}

	return a.db.ListUsers(ctx, opts)
}

func (a *Application) SyncUserMemberships(ctx context.Context, user *otf.User, orgs []string, teams []*otf.Team) (*otf.User, error) {
	if err := user.SyncMemberships(ctx, a.db, orgs, teams); err != nil {
		return nil, err
	}

	a.V(1).Info("synchronised user's memberships", "username", user.Username())

	return user, nil
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
