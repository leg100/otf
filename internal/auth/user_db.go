package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

// CreateUser persists a User to the DB.
func (db *pgdb) CreateUser(ctx context.Context, user *User) error {
	return db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertUser(ctx, pggen.InsertUserParams{
			ID:        sql.String(user.ID),
			Username:  sql.String(user.Username),
			CreatedAt: sql.Timestamptz(user.CreatedAt),
			UpdatedAt: sql.Timestamptz(user.UpdatedAt),
		})
		if err != nil {
			return sql.Error(err)
		}
		for _, team := range user.Teams {
			_, err = q.InsertTeamMembership(ctx, []string{user.Username}, sql.String(team.ID))
			if err != nil {
				return sql.Error(err)
			}
		}
		return nil
	})
}

func (db *pgdb) listUsers(ctx context.Context) ([]*User, error) {
	result, err := db.Conn(ctx).FindUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*User, len(result))
	for i, r := range result {
		users[i] = userRow(r).toUser()
	}
	return users, nil
}

func (db *pgdb) listOrganizationUsers(ctx context.Context, organization string) ([]*User, error) {
	result, err := db.Conn(ctx).FindUsersByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, err
	}
	users := make([]*User, len(result))
	for i, r := range result {
		users[i] = userRow(r).toUser()
	}
	return users, nil
}

func (db *pgdb) listTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	result, err := db.Conn(ctx).FindUsersByTeamID(ctx, sql.String(teamID))
	if err != nil {
		return nil, err
	}

	items := make([]*User, len(result))
	for i, r := range result {
		items[i] = userRow(r).toUser()
	}
	return items, nil
}

// getUser retrieves a user from the DB, along with its sessions.
func (db *pgdb) getUser(ctx context.Context, spec UserSpec) (*User, error) {
	if spec.UserID != nil {
		result, err := db.Conn(ctx).FindUserByID(ctx, sql.String(*spec.UserID))
		if err != nil {
			return nil, err
		}
		return userRow(result).toUser(), nil
	} else if spec.Username != nil {
		result, err := db.Conn(ctx).FindUserByUsername(ctx, sql.String(*spec.Username))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
	} else if spec.AuthenticationTokenID != nil {
		result, err := db.Conn(ctx).FindUserByAuthenticationTokenID(ctx, sql.String(*spec.AuthenticationTokenID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db *pgdb) addTeamMembership(ctx context.Context, teamID string, usernames ...string) error {
	_, err := db.Conn(ctx).InsertTeamMembership(ctx, usernames, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) removeTeamMembership(ctx context.Context, teamID string, usernames ...string) error {
	_, err := db.Conn(ctx).DeleteTeamMembership(ctx, usernames, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// DeleteUser deletes a user from the DB.
func (db *pgdb) DeleteUser(ctx context.Context, spec UserSpec) error {
	if spec.UserID != nil {
		_, err := db.Conn(ctx).DeleteUserByID(ctx, sql.String(*spec.UserID))
		if err != nil {
			return sql.Error(err)
		}
	} else if spec.Username != nil {
		_, err := db.Conn(ctx).DeleteUserByUsername(ctx, sql.String(*spec.Username))
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
	err = db.Tx(ctx, func(ctx context.Context, q pggen.Querier) (err error) {
		// First demote any existing site admins...
		resetted, err = q.ResetUserSiteAdmins(ctx)
		if err != nil {
			return err
		}
		// ...then promote any specified usernames
		if len(usernames) > 0 {
			updated, err = q.UpdateUserSiteAdmins(ctx, usernames)
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

//
// User tokens
//

func (db *pgdb) createUserToken(ctx context.Context, token *UserToken) error {
	_, err := db.Conn(ctx).InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     sql.String(token.ID),
		Description: sql.String(token.Description),
		Username:    sql.String(token.Username),
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
	})
	return err
}

func (db *pgdb) listUserTokens(ctx context.Context, username string) ([]*UserToken, error) {
	result, err := db.Conn(ctx).FindTokensByUsername(ctx, sql.String(username))
	if err != nil {
		return nil, err
	}
	tokens := make([]*UserToken, len(result))
	for i, row := range result {
		tokens[i] = &UserToken{
			ID:          row.TokenID.String,
			CreatedAt:   row.CreatedAt.Time.UTC(),
			Description: row.Description.String,
			Username:    row.Username.String,
		}
	}
	return tokens, nil
}

func (db *pgdb) getUserToken(ctx context.Context, id string) (*UserToken, error) {
	row, err := db.Conn(ctx).FindTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return &UserToken{
		ID:          row.TokenID.String,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Description: row.Description.String,
		Username:    row.Username.String,
	}, nil
}

func (db *pgdb) deleteUserToken(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
