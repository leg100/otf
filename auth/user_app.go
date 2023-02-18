package auth

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/rbac"
)

type userApp interface {
	CreateUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error)
	GetUser(ctx context.Context, spec otf.UserSpec) (otf.User, error)
	DeleteUser(ctx context.Context, username string) error
	AddOrganizationMembership(ctx context.Context, userID, organization string) error
	RemoveOrganizationMembership(ctx context.Context, userID, organization string) error

	createUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error)
	listUsers(ctx context.Context, organization string) ([]*User, error)
	getUser(ctx context.Context, spec otf.UserSpec) (*User, error)
	deleteUser(ctx context.Context, userID string) error

	addOrganizationMembership(ctx context.Context, userID, organization string) error
	removeOrganizationMembership(ctx context.Context, userID, organization string) error
	addTeamMembership(ctx context.Context, userID, teamID string) error
	removeTeamMembership(ctx context.Context, userID, teamID string) error

	sync(ctx context.Context, from cloud.User) (*User, error)
}

func (a *app) CreateUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error) {
	return a.createUser(ctx, username, opts...)
}

func (a *app) GetUser(ctx context.Context, spec otf.UserSpec) (otf.User, error) {
	user, err := a.db.getUser(ctx, spec)
	if err != nil {
		a.V(2).Info("retrieving user", "spec", spec)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username())

	return user, nil
}

func (a *app) AddOrganizationMembership(ctx context.Context, userID, organization string) error {
	return a.addOrganizationMembership(ctx, userID, organization)
}

func (a *app) RemoveOrganizationMembership(ctx context.Context, userID, organization string) error {
	return a.removeOrganizationMembership(ctx, userID, organization)
}

func (a *app) DeleteUser(ctx context.Context, userID string) error {
	return a.deleteUser(ctx, userID)
}

func (a *app) createUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error) {
	user := NewUser(username, opts...)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username)
		return nil, err
	}

	a.V(0).Info("created user", "username", username)

	return user, nil
}

func (a *app) getUser(ctx context.Context, spec otf.UserSpec) (*User, error) {
	user, err := a.db.getUser(ctx, spec)
	if err != nil {
		a.V(2).Info("retrieving user", "spec", spec)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username())

	return user, nil
}

// listUsers lists an organization's users
func (a *app) listUsers(ctx context.Context, organization string) ([]*User, error) {
	_, err := a.CanAccessOrganization(ctx, rbac.ListUsersAction, organization)
	if err != nil {
		return nil, err
	}

	return a.db.listUsers(ctx, organization)
}

func (a *app) addOrganizationMembership(ctx context.Context, userID, organization string) error {
	if err := a.db.addOrganizationMembership(ctx, userID, organization); err != nil {
		a.Error(err, "adding organization membership", "user", userID, "org", organization)
		return err
	}
	a.V(0).Info("added organization membership", "user", userID, "org", organization)

	return nil
}

func (a *app) removeOrganizationMembership(ctx context.Context, userID, organization string) error {
	if err := a.db.removeOrganizationMembership(ctx, userID, organization); err != nil {
		a.Error(err, "removing organization membership", "user", userID, "org", organization)
		return err
	}
	a.V(0).Info("removed organization membership", "user", userID, "org", organization)

	return nil
}

func (a *app) addTeamMembership(ctx context.Context, userID, teamID string) error {
	if err := a.db.addTeamMembership(ctx, userID, teamID); err != nil {
		a.Error(err, "adding team membership", "user", userID, "team", teamID)
		return err
	}
	a.V(0).Info("added team membership", "user", userID, "team", teamID)

	return nil
}

func (a *app) removeTeamMembership(ctx context.Context, userID, teamID string) error {
	if err := a.db.removeTeamMembership(ctx, userID, teamID); err != nil {
		a.Error(err, "removing team membership", "user", userID, "team", teamID)
		return err
	}
	a.V(0).Info("removed team membership", "user", userID, "team", teamID)

	return nil
}

func (a *app) deleteUser(ctx context.Context, userID string) error {
	err := a.db.DeleteUser(ctx, otf.UserSpec{UserID: otf.String(userID)})
	if err != nil {
		a.V(2).Info("deleting user", "id", userID)
		return err
	}

	a.V(2).Info("deleted user", "id", userID)

	return nil
}
