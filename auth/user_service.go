package auth

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/rbac"
)

type UserService interface {
	CreateUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error)
	GetUser(ctx context.Context, spec UserSpec) (*User, error)
	ListUsers(ctx context.Context, organization string) ([]*User, error)
	DeleteUser(ctx context.Context, username string) error
	AddOrganizationMembership(ctx context.Context, userID, organization string) error
	RemoveOrganizationMembership(ctx context.Context, userID, organization string) error
	AddTeamMembership(ctx context.Context, userID, teamID string) error
	RemoveTeamMembership(ctx context.Context, userID, teamID string) error

	sync(ctx context.Context, from cloud.User) (*User, error)
}

func (a *service) CreateUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error) {
	user := NewUser(username, opts...)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username)
		return nil, err
	}

	a.V(0).Info("created user", "username", username)

	return user, nil
}

func (a *service) GetUser(ctx context.Context, spec UserSpec) (*User, error) {
	user, err := a.db.getUser(ctx, spec)
	if err != nil {
		a.V(2).Info("retrieving user", "spec", spec)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username)

	return user, nil
}

// ListUsers lists an organization's users
func (a *service) ListUsers(ctx context.Context, organization string) ([]*User, error) {
	_, err := a.organization.CanAccess(ctx, rbac.ListUsersAction, organization)
	if err != nil {
		return nil, err
	}

	return a.db.listUsers(ctx, organization)
}

func (a *service) AddOrganizationMembership(ctx context.Context, userID, organization string) error {
	if err := a.db.addOrganizationMembership(ctx, userID, organization); err != nil {
		a.Error(err, "adding organization membership", "user", userID, "org", organization)
		return err
	}
	a.V(0).Info("added organization membership", "user", userID, "org", organization)

	return nil
}

func (a *service) RemoveOrganizationMembership(ctx context.Context, userID, organization string) error {
	if err := a.db.removeOrganizationMembership(ctx, userID, organization); err != nil {
		a.Error(err, "removing organization membership", "user", userID, "org", organization)
		return err
	}
	a.V(0).Info("removed organization membership", "user", userID, "org", organization)

	return nil
}

func (a *service) DeleteUser(ctx context.Context, username string) error {
	err := a.db.DeleteUser(ctx, UserSpec{Username: otf.String(username)})
	if err != nil {
		a.V(2).Info("deleting user", "username", username)
		return err
	}

	a.V(2).Info("deleted user", "username", username)

	return nil
}

func (a *service) AddTeamMembership(ctx context.Context, userID, teamID string) error {
	if err := a.db.addTeamMembership(ctx, userID, teamID); err != nil {
		a.Error(err, "adding team membership", "user", userID, "team", teamID)
		return err
	}
	a.V(0).Info("added team membership", "user", userID, "team", teamID)

	return nil
}

func (a *service) RemoveTeamMembership(ctx context.Context, userID, teamID string) error {
	if err := a.db.removeTeamMembership(ctx, userID, teamID); err != nil {
		a.Error(err, "removing team membership", "user", userID, "team", teamID)
		return err
	}
	a.V(0).Info("removed team membership", "user", userID, "team", teamID)

	return nil
}
