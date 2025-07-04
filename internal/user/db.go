package user

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/team"
)

type pgdb struct {
	*sql.DB // provides access to generated SQL queries
	logr.Logger
}

// CreateUser persists a User to the DB.
func (db *pgdb) CreateUser(ctx context.Context, user *User) error {
	return db.Tx(ctx, func(ctx context.Context) error {
		_, err := db.Exec(ctx, `
INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username,
	avatar_url
) VALUES (
    @user_id,
    @created_at,
    @updated_at,
    @username,
	@avatar_url
)
`,
			pgx.NamedArgs{
				"user_id":    user.ID,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
				"username":   user.Username,
				"avatar_url": user.AvatarURL,
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
`, team.ID, []Username{user.Username})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) updateAvatarURL(ctx context.Context, username Username, avatarURL string) error {
	_, err := db.Exec(ctx, `
UPDATE users
SET avatar_url = $1
WHERE username = $2
`, avatarURL, username)
	return err
}

func (db *pgdb) listUsers(ctx context.Context) ([]*User, error) {
	rows := db.Query(ctx, `
SELECT
    u.*,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
`)
	return sql.CollectRows(rows, scan)
}

func (db *pgdb) listOrganizationUsers(ctx context.Context, organization organization.Name) ([]*User, error) {
	rows := db.Query(ctx, `
SELECT
    u.*,
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
	return sql.CollectRows(rows, scan)
}

func (db *pgdb) listTeamUsers(ctx context.Context, teamID resource.TfeID) ([]*User, error) {
	rows := db.Query(ctx, `
SELECT
    u.*,
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
	return sql.CollectRows(rows, scan)
}

// getUser retrieves a user from the DB, along with its sessions.
func (db *pgdb) getUser(ctx context.Context, spec UserSpec) (*User, error) {
	var rows pgx.Rows
	if spec.UserID != nil {
		rows = db.Query(ctx, `
SELECT
    u.*,
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
    u.*,
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
    u.*,
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
	user, err := sql.CollectOneRow(rows, scan)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *pgdb) addTeamMembership(ctx context.Context, teamID resource.TfeID, usernames ...Username) error {
	_, err := db.Exec(ctx, `
WITH
    users AS (
        SELECT username
        FROM unnest($2::text[]) t(username)
    )
INSERT INTO team_memberships (username, team_id)
SELECT username, $1
FROM users
`, teamID, usernames)
	return err
}

func (db *pgdb) removeTeamMembership(ctx context.Context, teamID resource.TfeID, usernames ...Username) error {
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
`, *spec.UserID)
		return err
	} else if spec.Username != nil {
		_, err := db.Exec(ctx, `
DELETE
FROM users
WHERE username = $1
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
	err = db.Tx(ctx, func(ctx context.Context) (err error) {
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

func scan(row pgx.CollectableRow) (*User, error) {
	type model struct {
		ID        resource.TfeID `db:"user_id"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
		SiteAdmin bool           `db:"site_admin"`
		AvatarURL *string        `db:"avatar_url"`
		Username  Username
		Teams     []team.Model
	}
	m, err := pgx.RowToAddrOfStructByName[model](row)
	if err != nil {
		return nil, err
	}
	user := &User{
		ID:        m.ID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		SiteAdmin: m.SiteAdmin,
		Username:  m.Username,
		AvatarURL: m.AvatarURL,
	}
	// Only allocate if there are any teams; tests for equality otherwise fail
	// comparing nil with an empty slice.
	if len(m.Teams) > 0 {
		user.Teams = make([]*team.Team, len(m.Teams))
		for i, model := range m.Teams {
			user.Teams[i] = model.ToTeam()
		}
	}
	return user, nil
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
`,
		token.ID,
		token.CreatedAt,
		token.Description,
		token.Username,
	)
	return err
}

func (db *pgdb) listUserTokens(ctx context.Context, username Username) ([]*UserToken, error) {
	rows := db.Query(ctx, `
SELECT token_id, created_at, description, username
FROM tokens
WHERE username = $1
`, username)
	return sql.CollectRows(rows, scanToken)
}

func (db *pgdb) getUserToken(ctx context.Context, id resource.TfeID) (*UserToken, error) {
	rows := db.Query(ctx, `
SELECT token_id, created_at, description, username
FROM tokens
WHERE token_id = $1
`, id)
	return sql.CollectOneRow(rows, scanToken)
}

func (db *pgdb) deleteUserToken(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM tokens
WHERE token_id = $1
`, id)
	return err
}

func scanToken(row pgx.CollectableRow) (*UserToken, error) {
	return pgx.RowToAddrOfStructByName[UserToken](row)
}
