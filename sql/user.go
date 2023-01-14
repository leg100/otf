package sql

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateUser persists a User to the DB.
func (db *DB) CreateUser(ctx context.Context, user *otf.User) error {
	return db.tx(ctx, func(tx *DB) error {
		_, err := tx.InsertUser(ctx, pggen.InsertUserParams{
			ID:        String(user.ID()),
			Username:  String(user.Username()),
			CreatedAt: Timestamptz(user.CreatedAt()),
			UpdatedAt: Timestamptz(user.UpdatedAt()),
		})
		if err != nil {
			return err
		}
		for _, org := range user.Organizations() {
			_, err = tx.InsertOrganizationMembership(ctx, String(user.ID()), String(org))
			if err != nil {
				return err
			}
		}
		for _, team := range user.Teams() {
			_, err = tx.InsertTeamMembership(ctx, String(user.ID()), String(team.ID()))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *DB) ListUsers(ctx context.Context, opts otf.UserListOptions) ([]*otf.User, error) {
	var users []*otf.User
	if opts.Organization != nil && opts.TeamName != nil {
		result, err := db.FindUsersByTeam(ctx, String(*opts.Organization), String(*opts.TeamName))
		if err != nil {
			return nil, err
		}
		for _, r := range result {
			users = append(users, otf.UnmarshalUserResult(otf.UserResult(r)))
		}
	} else if opts.Organization != nil {
		result, err := db.FindUsersByOrganization(ctx, String(*opts.Organization))
		if err != nil {
			return nil, err
		}
		for _, r := range result {
			users = append(users, otf.UnmarshalUserResult(otf.UserResult(r)))
		}
	} else {
		result, err := db.FindUsers(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range result {
			users = append(users, otf.UnmarshalUserResult(otf.UserResult(r)))
		}
	}
	return users, nil
}

// GetUser retrieves a user from the DB, along with its sessions.
func (db *DB) GetUser(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	if spec.UserID != nil {
		result, err := db.FindUserByID(ctx, String(*spec.UserID))
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalUserResult(otf.UserResult(result)), nil
	} else if spec.Username != nil {
		result, err := db.FindUserByUsername(ctx, String(*spec.Username))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserResult(otf.UserResult(result)), nil
	} else if spec.AuthenticationToken != nil {
		result, err := db.FindUserByAuthenticationToken(ctx, String(*spec.AuthenticationToken))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserResult(otf.UserResult(result)), nil
	} else if spec.SessionToken != nil {
		result, err := db.FindUserBySessionToken(ctx, String(*spec.SessionToken))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserResult(otf.UserResult(result)), nil
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db *DB) AddOrganizationMembership(ctx context.Context, id, orgID string) error {
	_, err := db.InsertOrganizationMembership(ctx, String(id), String(orgID))
	if err != nil {
		return databaseError(err)
	}
	return nil
}

func (db *DB) RemoveOrganizationMembership(ctx context.Context, id, orgID string) error {
	_, err := db.DeleteOrganizationMembership(ctx, String(id), String(orgID))
	if err != nil {
		return databaseError(err)
	}
	return nil
}

// DeleteUser deletes a user from the DB.
func (db *DB) DeleteUser(ctx context.Context, spec otf.UserSpec) error {
	if spec.UserID != nil {
		_, err := db.DeleteUserByID(ctx, String(*spec.UserID))
		if err != nil {
			return databaseError(err)
		}
	} else if spec.Username != nil {
		_, err := db.DeleteUserByUsername(ctx, String(*spec.Username))
		if err != nil {
			return databaseError(err)
		}
	} else {
		return fmt.Errorf("unsupported user spec for deletion")
	}
	return nil
}
