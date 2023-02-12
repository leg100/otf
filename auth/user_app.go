package auth

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/rbac"
)

type userApp interface {
	GetUser(ctx context.Context, spec otf.UserSpec) (otf.User, error)

	createUser(ctx context.Context, username string) (*User, error)
	listUsers(ctx context.Context, organization string) ([]*User, error)
	getUser(ctx context.Context, spec otf.UserSpec) (*User, error)

	addOrganizationMembership(ctx context.Context, userID, organization string) error
	removeOrganizationMembership(ctx context.Context, userID, organization string) error
	addTeamMembership(ctx context.Context, userID, teamID string) error
	removeTeamMembership(ctx context.Context, userID, teamID string) error

	sync(ctx context.Context, from cloud.User) (*User, error)
}

func (a *Application) GetUser(ctx context.Context, spec otf.UserSpec) (otf.User, error) {
	user, err := a.db.getUser(ctx, spec)
	if err != nil {
		a.V(2).Info("retrieving user", "spec", spec)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username())

	return user, nil
}

func (a *Application) createUser(ctx context.Context, username string) (*User, error) {
	user := newUser(username)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username)
		return nil, err
	}

	a.V(0).Info("created user", "username", username)

	return user, nil
}

func (a *Application) getUser(ctx context.Context, spec otf.UserSpec) (*User, error) {
	user, err := a.db.getUser(ctx, spec)
	if err != nil {
		a.V(2).Info("retrieving user", "spec", spec)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username())

	return user, nil
}

// listUsers lists an organization's users
func (a *Application) listUsers(ctx context.Context, organization string) ([]*User, error) {
	_, err := a.CanAccessOrganization(ctx, rbac.ListUsersAction, organization)
	if err != nil {
		return nil, err
	}

	return a.db.listUsers(ctx, organization)
}

func (a *Application) addOrganizationMembership(ctx context.Context, userID, organization string) error {
	if err := a.db.addOrganizationMembership(ctx, userID, organization); err != nil {
		a.Error(err, "adding organization membership", "user", userID, "org", organization)
		return err
	}
	a.V(0).Info("added organization membership", "user", userID, "org", organization)

	return nil
}

func (a *Application) removeOrganizationMembership(ctx context.Context, userID, organization string) error {
	if err := a.db.removeOrganizationMembership(ctx, userID, organization); err != nil {
		a.Error(err, "removing organization membership", "user", userID, "org", organization)
		return err
	}
	a.V(0).Info("removed organization membership", "user", userID, "org", organization)

	return nil
}

func (a *Application) addTeamMembership(ctx context.Context, userID, teamID string) error {
	if err := a.db.addTeamMembership(ctx, userID, teamID); err != nil {
		a.Error(err, "adding team membership", "user", userID, "team", teamID)
		return err
	}
	a.V(0).Info("added team membership", "user", userID, "team", teamID)

	return nil
}

func (a *Application) removeTeamMembership(ctx context.Context, userID, teamID string) error {
	if err := a.db.removeTeamMembership(ctx, userID, teamID); err != nil {
		a.Error(err, "removing team membership", "user", userID, "team", teamID)
		return err
	}
	a.V(0).Info("removed team membership", "user", userID, "team", teamID)

	return nil
}
