package auth

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// CreateUser persists a User to the DB.
func (db *pgdb) CreateUser(ctx context.Context, user *User) error {
	return db.Transaction(ctx, func(tx otf.Database) error {
		_, err := tx.InsertUser(ctx, pggen.InsertUserParams{
			ID:        sql.String(user.ID()),
			Username:  sql.String(user.Username()),
			CreatedAt: sql.Timestamptz(user.CreatedAt()),
			UpdatedAt: sql.Timestamptz(user.UpdatedAt()),
		})
		if err != nil {
			return err
		}
		for _, org := range user.Organizations() {
			_, err = tx.InsertOrganizationMembership(ctx, sql.String(user.ID()), sql.String(org))
			if err != nil {
				return err
			}
		}
		for _, team := range user.teams {
			_, err = tx.InsertTeamMembership(ctx, sql.String(user.ID()), sql.String(team.ID()))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) listUsers(ctx context.Context, organization string) ([]*User, error) {
	result, err := db.FindUsersByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, err
	}

	var users []*User
	for _, r := range result {
		users = append(users, userRow(r).toUser())
	}
	return users, nil
}

func (db *pgdb) listTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	result, err := db.FindUsersByTeamID(ctx, sql.String(teamID))
	if err != nil {
		return nil, err
	}

	var items []*User
	for _, r := range result {
		items = append(items, userRow(r).toUser())
	}
	return items, nil
}

// getUser retrieves a user from the DB, along with its sessions.
func (db *pgdb) getUser(ctx context.Context, spec otf.UserSpec) (*User, error) {
	if spec.UserID != nil {
		result, err := db.FindUserByID(ctx, sql.String(*spec.UserID))
		if err != nil {
			return nil, err
		}
		return userRow(result).toUser(), nil
	} else if spec.Username != nil {
		result, err := db.FindUserByUsername(ctx, sql.String(*spec.Username))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
	} else if spec.AuthenticationToken != nil {
		result, err := db.FindUserByAuthenticationToken(ctx, sql.String(*spec.AuthenticationToken))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
	} else if spec.SessionToken != nil {
		result, err := db.FindUserBySessionToken(ctx, sql.String(*spec.SessionToken))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db *pgdb) addOrganizationMembership(ctx context.Context, userID, organization string) error {
	_, err := db.InsertOrganizationMembership(ctx, sql.String(userID), sql.String(organization))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) removeOrganizationMembership(ctx context.Context, id, orgID string) error {
	_, err := db.DeleteOrganizationMembership(ctx, sql.String(id), sql.String(orgID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) addTeamMembership(ctx context.Context, userID, teamID string) error {
	_, err := db.InsertTeamMembership(ctx, sql.String(userID), sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) removeTeamMembership(ctx context.Context, userID, teamID string) error {
	_, err := db.DeleteTeamMembership(ctx, sql.String(userID), sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// DeleteUser deletes a user from the DB.
func (db *pgdb) DeleteUser(ctx context.Context, spec otf.UserSpec) error {
	if spec.UserID != nil {
		_, err := db.DeleteUserByID(ctx, sql.String(*spec.UserID))
		if err != nil {
			return sql.Error(err)
		}
	} else if spec.Username != nil {
		_, err := db.DeleteUserByUsername(ctx, sql.String(*spec.Username))
		if err != nil {
			return sql.Error(err)
		}
	} else {
		return fmt.Errorf("unsupported user spec for deletion")
	}
	return nil
}
