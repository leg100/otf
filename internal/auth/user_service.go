package auth

import (
	"context"
	"errors"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql/pggen"
)

var ErrCannotDeleteOnlyOwner = errors.New("cannot remove the last owner")

type (
	UserService interface {
		CreateUser(ctx context.Context, username string, opts ...NewUserOption) (*User, error)
		GetUser(ctx context.Context, spec UserSpec) (*User, error)
		ListUsers(ctx context.Context) ([]*User, error)
		ListOrganizationUsers(ctx context.Context, organization string) ([]*User, error)
		DeleteUser(ctx context.Context, username string) error
		AddTeamMembership(ctx context.Context, teamID string, usernames []string) error
		RemoveTeamMembership(ctx context.Context, teamID string, usernames []string) error
		SetSiteAdmins(ctx context.Context, usernames ...string) error
	}
)

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
	subject, err := a.site.CanAccess(ctx, rbac.GetUserAction, "")
	if err != nil {
		return nil, err
	}

	user, err := a.db.getUser(ctx, spec)
	if err != nil {
		a.V(9).Info("retrieving user", "spec", spec, "subject", subject)
		return nil, err
	}

	a.V(9).Info("retrieved user", "username", user.Username, "subject", subject)

	return user, nil
}

// ListUsers lists all users.
func (a *service) ListUsers(ctx context.Context) ([]*User, error) {
	_, err := a.site.CanAccess(ctx, rbac.ListUsersAction, "")
	if err != nil {
		return nil, err
	}

	return a.db.listUsers(ctx)
}

// ListUsers lists an organization's users
func (a *service) ListOrganizationUsers(ctx context.Context, organization string) ([]*User, error) {
	_, err := a.organization.CanAccess(ctx, rbac.ListUsersAction, organization)
	if err != nil {
		return nil, err
	}

	return a.db.listOrganizationUsers(ctx, organization)
}

func (a *service) DeleteUser(ctx context.Context, username string) error {
	subject, err := a.site.CanAccess(ctx, rbac.DeleteUserAction, "")
	if err != nil {
		return err
	}

	err = a.db.DeleteUser(ctx, UserSpec{Username: internal.String(username)})
	if err != nil {
		a.Error(err, "deleting user", "username", username, "subject", subject)
		return err
	}

	a.V(2).Info("deleted user", "username", username, "subject", subject)

	return nil
}

// AddTeamMembership adds users to a team. If a user does not exist then the
// user is created first.
func (a *service) AddTeamMembership(ctx context.Context, teamID string, usernames []string) error {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.AddTeamMembershipAction, team.Organization)
	if err != nil {
		return err
	}

	err = a.db.Tx(ctx, func(ctx context.Context, _ pggen.Querier) error {
		// Check each username: if user does not exist then create user.
		for _, username := range usernames {
			_, err := a.db.getUser(ctx, UserSpec{Username: &username})
			if errors.Is(err, internal.ErrResourceNotFound) {
				if _, err := a.CreateUser(ctx, username); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		if err := a.db.addTeamMembership(ctx, teamID, usernames...); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		a.Error(err, "adding team membership", "user", usernames, "team", teamID, "subject", subject)
		return err
	}

	a.V(0).Info("added team membership", "users", usernames, "team", teamID, "subject", subject)

	return nil
}

// RemoveTeamMembership removes users from a team.
func (a *service) RemoveTeamMembership(ctx context.Context, teamID string, usernames []string) error {
	team, err := a.db.getTeamForUpdate(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.RemoveTeamMembershipAction, team.Organization)
	if err != nil {
		return err
	}

	// check whether all members of the owners group are going to be deleted
	// (which is not allowed)
	if team.Name == "owners" {
		if owners, err := a.db.listTeamMembers(ctx, team.ID); err != nil {
			a.Error(err, "removing team membership: listing team members", "team_id", team.ID, "subject", subject)
			return err
		} else if len(owners) <= len(usernames) {
			return ErrCannotDeleteOnlyOwner
		}
	}

	if err := a.db.removeTeamMembership(ctx, teamID, usernames...); err != nil {
		a.Error(err, "removing team membership", "users", usernames, "team", teamID, "subject", subject)
		return err
	}
	a.V(0).Info("removed team membership", "users", usernames, "team", teamID, "subject", subject)

	return nil
}

// SetSiteAdmins authoritatively promotes users with the given usernames to site
// admins. If no such users exist then they are created. Any unspecified users
// that are currently site admins are demoted.
func (a *service) SetSiteAdmins(ctx context.Context, usernames ...string) error {
	for _, username := range usernames {
		_, err := a.db.getUser(ctx, UserSpec{Username: &username})
		if err == internal.ErrResourceNotFound {
			if _, err = a.CreateUser(ctx, username); err != nil {
				return err
			}
		}
	}
	promoted, demoted, err := a.db.setSiteAdmins(ctx, usernames...)
	if err != nil {
		a.Error(err, "setting site admins", "users", usernames)
		return err
	}
	a.V(0).Info("set site admins", "admins", usernames, "promoted", promoted, "demoted", demoted)
	return nil
}
