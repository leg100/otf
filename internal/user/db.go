package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/team"
)

// dbresult represents the result of a database query for a user.
type dbresult struct {
	UserID    resource.ID
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
	Teams     []sqlc.Team
}

func (result dbresult) toUser() *User {
	user := User{
		ID:        result.UserID,
		CreatedAt: result.CreatedAt.Time.UTC(),
		UpdatedAt: result.UpdatedAt.Time.UTC(),
		Username:  result.Username.String,
		SiteAdmin: result.SiteAdmin.Bool,
	}
	for _, tr := range result.Teams {
		user.Teams = append(user.Teams, team.TeamRow(tr).ToTeam())
	}
	return &user
}

// pgdb stores user resources in a postgres database
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
	logr.Logger
}

// CreateUser persists a User to the DB.
func (db *pgdb) CreateUser(ctx context.Context, user *User) error {
	return db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		err := q.InsertUser(ctx, sqlc.InsertUserParams{
			ID:        user.ID,
			Username:  sql.String(user.Username),
			CreatedAt: sql.Timestamptz(user.CreatedAt),
			UpdatedAt: sql.Timestamptz(user.UpdatedAt),
		})
		if err != nil {
			return sql.Error(err)
		}
		for _, team := range user.Teams {
			_, err = q.InsertTeamMembership(ctx, sqlc.InsertTeamMembershipParams{
				TeamID:    team.ID,
				Usernames: sql.StringArray([]string{user.Username}),
			})
			if err != nil {
				return sql.Error(err)
			}
		}
		return nil
	})
}

func (db *pgdb) listUsers(ctx context.Context) ([]*User, error) {
	result, err := db.Querier(ctx).FindUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*User, len(result))
	for i, r := range result {
		users[i] = dbresult(r).toUser()
	}
	return users, nil
}

func (db *pgdb) listOrganizationUsers(ctx context.Context, organization string) ([]*User, error) {
	result, err := db.Querier(ctx).FindUsersByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, err
	}
	users := make([]*User, len(result))
	for i, r := range result {
		users[i] = dbresult(r).toUser()
	}
	return users, nil
}

func (db *pgdb) listTeamUsers(ctx context.Context, teamID resource.ID) ([]*User, error) {
	result, err := db.Querier(ctx).FindUsersByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	items := make([]*User, len(result))
	for i, r := range result {
		items[i] = dbresult(r).toUser()
	}
	return items, nil
}

// getUser retrieves a user from the DB, along with its sessions.
func (db *pgdb) getUser(ctx context.Context, spec UserSpec) (*User, error) {
	if spec.UserID != nil {
		result, err := db.Querier(ctx).FindUserByID(ctx, *spec.UserID)
		if err != nil {
			return nil, err
		}
		return dbresult(result).toUser(), nil
	} else if spec.Username != nil {
		result, err := db.Querier(ctx).FindUserByUsername(ctx, sql.String(*spec.Username))
		if err != nil {
			return nil, sql.Error(err)
		}
		return dbresult(result).toUser(), nil
	} else if spec.AuthenticationTokenID != nil {
		result, err := db.Querier(ctx).FindUserByAuthenticationTokenID(ctx, *spec.AuthenticationTokenID)
		if err != nil {
			return nil, sql.Error(err)
		}
		return dbresult(result).toUser(), nil
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db *pgdb) addTeamMembership(ctx context.Context, teamID resource.ID, usernames ...string) error {
	_, err := db.Querier(ctx).InsertTeamMembership(ctx, sqlc.InsertTeamMembershipParams{
		Usernames: sql.StringArray(usernames),
		TeamID:    teamID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) removeTeamMembership(ctx context.Context, teamID resource.ID, usernames ...string) error {
	_, err := db.Querier(ctx).DeleteTeamMembership(ctx, sqlc.DeleteTeamMembershipParams{
		Usernames: sql.StringArray(usernames),
		TeamID:    teamID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// DeleteUser deletes a user from the DB.
func (db *pgdb) DeleteUser(ctx context.Context, spec UserSpec) error {
	if spec.UserID != nil {
		_, err := db.Querier(ctx).DeleteUserByID(ctx, *spec.UserID)
		if err != nil {
			return sql.Error(err)
		}
	} else if spec.Username != nil {
		_, err := db.Querier(ctx).DeleteUserByUsername(ctx, sql.String(*spec.Username))
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
	err = db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) (err error) {
		// First demote any existing site admins...
		resetted, err = q.ResetUserSiteAdmins(ctx)
		if err != nil {
			return err
		}
		// ...then promote any specified usernames
		if len(usernames) > 0 {
			updated, err = q.UpdateUserSiteAdmins(ctx, sql.StringArray(usernames))
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
	err := db.Querier(ctx).InsertToken(ctx, sqlc.InsertTokenParams{
		TokenID:     token.ID,
		Description: sql.String(token.Description),
		Username:    sql.String(token.Username),
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
	})
	return err
}

func (db *pgdb) listUserTokens(ctx context.Context, username string) ([]*UserToken, error) {
	result, err := db.Querier(ctx).FindTokensByUsername(ctx, sql.String(username))
	if err != nil {
		return nil, err
	}
	tokens := make([]*UserToken, len(result))
	for i, row := range result {
		tokens[i] = &UserToken{
			ID:          row.TokenID,
			CreatedAt:   row.CreatedAt.Time.UTC(),
			Description: row.Description.String,
			Username:    row.Username.String,
		}
	}
	return tokens, nil
}

func (db *pgdb) getUserToken(ctx context.Context, id resource.ID) (*UserToken, error) {
	row, err := db.Querier(ctx).FindTokenByID(ctx, id)
	if err != nil {
		return nil, sql.Error(err)
	}
	return &UserToken{
		ID:          row.TokenID,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Description: row.Description.String,
		Username:    row.Username.String,
	}, nil
}

func (db *pgdb) deleteUserToken(ctx context.Context, id resource.ID) error {
	_, err := db.Querier(ctx).DeleteTokenByID(ctx, id)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
