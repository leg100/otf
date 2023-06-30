package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

// CreateUser persists a User to the DB.
func (db *pgdb) CreateUser(ctx context.Context, user *User) error {
	return db.Tx(ctx, func(tx internal.DB) error {
		_, err := tx.InsertUser(ctx, pggen.InsertUserParams{
			ID:        sql.String(user.ID),
			Username:  sql.String(user.Username),
			CreatedAt: sql.Timestamptz(user.CreatedAt),
			UpdatedAt: sql.Timestamptz(user.UpdatedAt),
		})
		if err != nil {
			return sql.Error(err)
		}
		for _, team := range user.Teams {
			_, err = tx.InsertTeamMembership(ctx, []string{user.Username}, sql.String(team.ID))
			if err != nil {
				return sql.Error(err)
			}
		}
		return nil
	})
}

func (db *pgdb) listUsers(ctx context.Context) ([]*User, error) {
	result, err := db.FindUsers(ctx)
	if err != nil {
		return nil, err
	}
	var users []*User
	for _, r := range result {
		users = append(users, userRow(r).toUser())
	}
	return users, nil
}

func (db *pgdb) listOrganizationUsers(ctx context.Context, organization string) ([]*User, error) {
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
func (db *pgdb) getUser(ctx context.Context, spec UserSpec) (*User, error) {
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
	} else if spec.AuthenticationTokenID != nil {
		result, err := db.FindUserByAuthenticationTokenID(ctx, sql.String(*spec.AuthenticationTokenID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db *pgdb) addTeamMembership(ctx context.Context, teamID string, usernames ...string) error {
	_, err := db.InsertTeamMembership(ctx, usernames, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) removeTeamMembership(ctx context.Context, teamID string, usernames ...string) error {
	_, err := db.DeleteTeamMembership(ctx, usernames, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// DeleteUser deletes a user from the DB.
func (db *pgdb) DeleteUser(ctx context.Context, spec UserSpec) error {
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

// setSiteAdmins authoritatively promotes the given users to site admins,
// demoting all other site admins. The list of newly promoted and demoted users
// is returned.
func (db *pgdb) setSiteAdmins(ctx context.Context, usernames ...string) (promoted []string, demoted []string, err error) {
	var resetted, updated []pgtype.Text
	err = db.Tx(ctx, func(tx internal.DB) (err error) {
		// First demote any existing site admins...
		resetted, err = tx.ResetUserSiteAdmins(ctx)
		if err != nil {
			return err
		}
		// ...then promote any specified usernames
		if len(usernames) > 0 {
			updated, err = tx.UpdateUserSiteAdmins(ctx, usernames)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return pgtextSliceDiff(updated, resetted), pgtextSliceDiff(resetted, updated), nil
}

// pgtextSliceDiff returns the elements in `a` that aren't in `b`.
func pgtextSliceDiff(a, b []pgtype.Text) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x.String] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x.String]; !found {
			diff = append(diff, x.String)
		}
	}
	return diff
}
