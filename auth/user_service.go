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
	AddTeamMembership(ctx context.Context, username, teamID string) error
	RemoveTeamMembership(ctx context.Context, username, teamID string) error

	sync(ctx context.Context, from cloud.User) (*User, error)
}

func (a *service) CreateUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error) {
	subject, err := a.site.CanAccess(ctx, rbac.CreateUserAction, "")
	if err != nil {
		return nil, err
	}

	user := NewUser(username, opts...)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username, "subject", subject)
		return nil, err
	}

	a.V(0).Info("created user", "username", username, "subject", subject)

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

func (a *service) DeleteUser(ctx context.Context, username string) error {
	subject, err := a.site.CanAccess(ctx, rbac.DeleteUserAction, "")
	if err != nil {
		return err
	}

	err = a.db.DeleteUser(ctx, UserSpec{Username: otf.String(username)})
	if err != nil {
		a.V(2).Info("deleting user", "username", username, "subject", subject)
		return err
	}

	a.V(2).Info("deleted user", "username", username, "subject", subject)

	return nil
}

func (a *service) AddTeamMembership(ctx context.Context, username, teamID string) error {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.AddTeamMembershipAction, team.Organization)
	if err != nil {
		return err
	}

	if err := a.db.addTeamMembership(ctx, username, teamID); err != nil {
		a.Error(err, "adding team membership", "user", username, "team", teamID, "subject", subject)
		return err
	}
	a.V(0).Info("added team membership", "user", username, "team", teamID, "subject", subject)

	return nil
}

func (a *service) RemoveTeamMembership(ctx context.Context, username, teamID string) error {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.RemoveTeamMembershipAction, team.Organization)
	if err != nil {
		return err
	}

	if err := a.db.removeTeamMembership(ctx, username, teamID); err != nil {
		a.Error(err, "removing team membership", "user", username, "team", teamID, "subject", subject)
		return err
	}
	a.V(0).Info("removed team membership", "user", username, "team", teamID, "subject", subject)

	return nil
}
