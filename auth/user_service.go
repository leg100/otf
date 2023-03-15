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
	DeleteUser(ctx context.Context, username string) error
	AddOrganizationMembership(ctx context.Context, userID, organization string) error
	RemoveOrganizationMembership(ctx context.Context, userID, organization string) error

	createUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error)
	listUsers(ctx context.Context, organization string) ([]*User, error)
	getUser(ctx context.Context, spec UserSpec) (*User, error)
	deleteUser(ctx context.Context, userID string) error

	addOrganizationMembership(ctx context.Context, userID, organization string) error
	removeOrganizationMembership(ctx context.Context, userID, organization string) error
	addTeamMembership(ctx context.Context, userID, teamID string) error
	removeTeamMembership(ctx context.Context, userID, teamID string) error

	sync(ctx context.Context, from cloud.User) (*User, error)
}

func (a *service) CreateUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error) {
	return a.createUser(ctx, username, opts...)
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

func (a *service) AddOrganizationMembership(ctx context.Context, userID, organization string) error {
	return a.addOrganizationMembership(ctx, userID, organization)
}

func (a *service) RemoveOrganizationMembership(ctx context.Context, userID, organization string) error {
	return a.removeOrganizationMembership(ctx, userID, organization)
}

func (a *service) DeleteUser(ctx context.Context, userID string) error {
	return a.deleteUser(ctx, userID)
}

func (a *service) createUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error) {
	user := NewUser(username, opts...)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username)
		return nil, err
	}

	a.V(0).Info("created user", "username", username)

	return user, nil
}

func (a *service) getUser(ctx context.Context, spec UserSpec) (*User, error) {
	user, err := a.db.getUser(ctx, spec)
	if err != nil {
		a.V(2).Info("retrieving user", "spec", spec)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username)

	return user, nil
}

// listUsers lists an organization's users
func (a *service) listUsers(ctx context.Context, organization string) ([]*User, error) {
	_, err := a.organization.CanAccess(ctx, rbac.ListUsersAction, organization)
	if err != nil {
		return nil, err
	}

	return a.db.listUsers(ctx, organization)
}

func (a *service) addOrganizationMembership(ctx context.Context, userID, organization string) error {
	if err := a.db.addOrganizationMembership(ctx, userID, organization); err != nil {
		a.Error(err, "adding organization membership", "user", userID, "org", organization)
		return err
	}
	a.V(0).Info("added organization membership", "user", userID, "org", organization)

	return nil
}

func (a *service) removeOrganizationMembership(ctx context.Context, userID, organization string) error {
	if err := a.db.removeOrganizationMembership(ctx, userID, organization); err != nil {
		a.Error(err, "removing organization membership", "user", userID, "org", organization)
		return err
	}
	a.V(0).Info("removed organization membership", "user", userID, "org", organization)

	return nil
}

func (a *service) addTeamMembership(ctx context.Context, userID, teamID string) error {
	if err := a.db.addTeamMembership(ctx, userID, teamID); err != nil {
		a.Error(err, "adding team membership", "user", userID, "team", teamID)
		return err
	}
	a.V(0).Info("added team membership", "user", userID, "team", teamID)

	return nil
}

func (a *service) removeTeamMembership(ctx context.Context, userID, teamID string) error {
	if err := a.db.removeTeamMembership(ctx, userID, teamID); err != nil {
		a.Error(err, "removing team membership", "user", userID, "team", teamID)
		return err
	}
	a.V(0).Info("removed team membership", "user", userID, "team", teamID)

	return nil
}

func (a *service) deleteUser(ctx context.Context, userID string) error {
	err := a.db.DeleteUser(ctx, UserSpec{UserID: otf.String(userID)})
	if err != nil {
		a.V(2).Info("deleting user", "id", userID)
		return err
	}

	a.V(2).Info("deleted user", "id", userID)

	return nil
}
