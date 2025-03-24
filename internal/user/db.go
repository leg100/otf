package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type pgdb struct {
	*sql.DB // provides access to generated SQL queries
	logr.Logger
}

// CreateUser persists a User to the DB.
func (db *pgdb) CreateUser(ctx context.Context, user *User) error {
	return db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		_, err := db.Exec(ctx, `
INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username
) VALUES (
    @user_id,
    @created_at,
    @updated_at,
    @username
)
`,
			pgx.NamedArgs{
				"id":         user.ID,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
				"username":   user.Username,
			},
		)
		if err != nil {
			return err
		}
		for _, team := range user.Teams {
			_, err := db.Exec(ctx, `
WITH
    users AS (
        SELECT username
        FROM unnest($2::text[]) t(username)
    )
INSERT INTO team_memberships (username, team_id)
SELECT username, $1
FROM users
RETURNING username
`, team.ID, []string{user.Username})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) listUsers(ctx context.Context) ([]*User, error) {
	rows := db.Query(ctx, `
SELECT
    u.user_id, u.username, u.created_at, u.updated_at, u.site_admin,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
`)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listOrganizationUsers(ctx context.Context, organization resource.OrganizationName) ([]*User, error) {
	rows := db.Query(ctx, `
SELECT
    u.user_id, u.username, u.created_at, u.updated_at, u.site_admin,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.organization_name = $1
GROUP BY u.user_id
`, organization)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listTeamUsers(ctx context.Context, teamID resource.TfeID) ([]*User, error) {
	rows := db.Query(ctx, `
SELECT
    u.user_id, u.username, u.created_at, u.updated_at, u.site_admin,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.team_id = $1
GROUP BY u.user_id
`, teamID)
	return sql.CollectRows(rows, db.scan)
}

// getUser retrieves a user from the DB, along with its sessions.
func (db *pgdb) getUser(ctx context.Context, spec UserSpec) (*User, error) {
	var rows pgx.Rows
	if spec.UserID != nil {
		rows = db.Query(ctx, `
SELECT
    u.user_id, u.username, u.created_at, u.updated_at, u.site_admin,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
WHERE u.user_id = $1
`, *spec.UserID)
	} else if spec.Username != nil {
		rows = db.Query(ctx, `
SELECT
    u.user_id, u.username, u.created_at, u.updated_at, u.site_admin,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
WHERE u.username = $1
`, *spec.Username)
	} else if spec.AuthenticationTokenID != nil {
		rows = db.Query(ctx, `
SELECT
    u.user_id, u.username, u.created_at, u.updated_at, u.site_admin,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
JOIN tokens t ON u.username = t.username
WHERE t.token_id = $1
`, *spec.AuthenticationTokenID)
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
	return sql.CollectOneRow(rows, db.scan)
}

func (db *pgdb) addTeamMembership(ctx context.Context, teamID resource.TfeID, usernames ...string) error {
	_, err := db.Exec(ctx, `
WITH
    users AS (
        SELECT username
        FROM unnest($2::text[]) t(username)
    )
INSERT INTO team_memberships (username, team_id)
SELECT username, $1
FROM users
RETURNING username
`, teamID, usernames)
	return err
}

func (db *pgdb) removeTeamMembership(ctx context.Context, teamID resource.TfeID, usernames ...string) error {
	_, err := db.Exec(ctx, `
WITH
    users AS (
        SELECT username
        FROM unnest($2::text[]) t(username)
    )
DELETE
FROM team_memberships tm
USING users
WHERE
    tm.username = users.username AND
    tm.team_id  = $1
RETURNING tm.username
`, teamID, usernames)
	return err
}

// DeleteUser deletes a user from the DB.
func (db *pgdb) DeleteUser(ctx context.Context, spec UserSpec) error {
	if spec.UserID != nil {
		_, err := db.Exec(ctx, `
DELETE
FROM users
WHERE user_id = $1
RETURNING user_id
`, *spec.UserID)
		return err
	} else if spec.Username != nil {
		_, err := db.Exec(ctx, `
DELETE
FROM users
WHERE username = $1
RETURNING user_id
`, *spec.Username)
		return err
	} else {
		return fmt.Errorf("unsupported user spec for deletion")
	}
}

// setSiteAdmins authoritatively promotes the given users to site admins,
// demoting all other site admins. The list of newly promoted and demoted users
// is returned.
func (db *pgdb) setSiteAdmins(ctx context.Context, usernames ...string) (promoted []string, demoted []string, err error) {
	var resetted, updated []string
	err = db.Tx(ctx, func(ctx context.Context, conn sql.Connection) (err error) {
		// First demote any existing site admins...
		rows := db.Query(ctx, `
UPDATE users
SET site_admin = false
WHERE site_admin = true
RETURNING username
`)
		resetted, err = sql.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			return err
		}
		// ...then promote any specified usernames
		if len(usernames) > 0 {
			rows := db.Query(ctx, `
UPDATE users
SET site_admin = true
WHERE username = ANY($1::text[])
RETURNING username
`, usernames)
			updated, err = sql.CollectRows(rows, pgx.RowTo[string])
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return internal.Diff(updated, resetted), internal.Diff(resetted, updated), nil
}

func (db *pgdb) scan(row pgx.CollectableRow) (*User, error) {
	user, err := pgx.RowToAddrOfStructByName[User](row)
	if err != nil {
		return nil, err
	}
	user.CreatedAt = user.CreatedAt.UTC()
	user.UpdatedAt = user.UpdatedAt.UTC()
	return user, nil
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
	_, err := db.Exec(ctx, `
INSERT INTO tokens (
    token_id,
    created_at,
    description,
    username
) VALUES (
    $1,
    $2,
    $3,
    $4
)
`
	err := q.InsertToken(ctx, db.Conn(ctx), InsertTokenParams{
		TokenID:     token.ID,
		Description: sql.String(token.Description),
		Username:    sql.String(token.Username),
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
	})
	return err
}

func (db *pgdb) listUserTokens(ctx context.Context, username string) ([]*UserToken, error) {
	result, err := q.FindTokensByUsername(ctx, db.Conn(ctx), sql.String(username))
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

func (db *pgdb) getUserToken(ctx context.Context, id resource.TfeID) (*UserToken, error) {
	row, err := q.FindTokenByID(ctx, db.Conn(ctx), id)
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

func (db *pgdb) deleteUserToken(ctx context.Context, id resource.TfeID) error {
	_, err := q.DeleteTokenByID(ctx, db.Conn(ctx), id)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
