// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: user.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

const deleteUserByID = `-- name: DeleteUserByID :one
DELETE
FROM users
WHERE user_id = $1
RETURNING user_id
`

func (q *Queries) DeleteUserByID(ctx context.Context, userID resource.ID) (resource.ID, error) {
	row := q.db.QueryRow(ctx, deleteUserByID, userID)
	var user_id resource.ID
	err := row.Scan(&user_id)
	return user_id, err
}

const deleteUserByUsername = `-- name: DeleteUserByUsername :one
DELETE
FROM users
WHERE username = $1
RETURNING user_id
`

func (q *Queries) DeleteUserByUsername(ctx context.Context, username pgtype.Text) (resource.ID, error) {
	row := q.db.QueryRow(ctx, deleteUserByUsername, username)
	var user_id resource.ID
	err := row.Scan(&user_id)
	return user_id, err
}

const findUserByAuthenticationTokenID = `-- name: FindUserByAuthenticationTokenID :one
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
`

type FindUserByAuthenticationTokenIDRow struct {
	UserID    resource.ID
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
	Teams     []Team
}

func (q *Queries) FindUserByAuthenticationTokenID(ctx context.Context, tokenID resource.ID) (FindUserByAuthenticationTokenIDRow, error) {
	row := q.db.QueryRow(ctx, findUserByAuthenticationTokenID, tokenID)
	var i FindUserByAuthenticationTokenIDRow
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.SiteAdmin,
		&i.Teams,
	)
	return i, err
}

const findUserByID = `-- name: FindUserByID :one
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
`

type FindUserByIDRow struct {
	UserID    resource.ID
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
	Teams     []Team
}

func (q *Queries) FindUserByID(ctx context.Context, userID resource.ID) (FindUserByIDRow, error) {
	row := q.db.QueryRow(ctx, findUserByID, userID)
	var i FindUserByIDRow
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.SiteAdmin,
		&i.Teams,
	)
	return i, err
}

const findUserByUsername = `-- name: FindUserByUsername :one
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
`

type FindUserByUsernameRow struct {
	UserID    resource.ID
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
	Teams     []Team
}

func (q *Queries) FindUserByUsername(ctx context.Context, username pgtype.Text) (FindUserByUsernameRow, error) {
	row := q.db.QueryRow(ctx, findUserByUsername, username)
	var i FindUserByUsernameRow
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.SiteAdmin,
		&i.Teams,
	)
	return i, err
}

const findUsers = `-- name: FindUsers :many
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
`

type FindUsersRow struct {
	UserID    resource.ID
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
	Teams     []Team
}

func (q *Queries) FindUsers(ctx context.Context) ([]FindUsersRow, error) {
	rows, err := q.db.Query(ctx, findUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindUsersRow
	for rows.Next() {
		var i FindUsersRow
		if err := rows.Scan(
			&i.UserID,
			&i.Username,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.SiteAdmin,
			&i.Teams,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findUsersByOrganization = `-- name: FindUsersByOrganization :many
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
`

type FindUsersByOrganizationRow struct {
	UserID    resource.ID
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
	Teams     []Team
}

func (q *Queries) FindUsersByOrganization(ctx context.Context, organizationName pgtype.Text) ([]FindUsersByOrganizationRow, error) {
	rows, err := q.db.Query(ctx, findUsersByOrganization, organizationName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindUsersByOrganizationRow
	for rows.Next() {
		var i FindUsersByOrganizationRow
		if err := rows.Scan(
			&i.UserID,
			&i.Username,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.SiteAdmin,
			&i.Teams,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findUsersByTeamID = `-- name: FindUsersByTeamID :many
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
`

type FindUsersByTeamIDRow struct {
	UserID    resource.ID
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
	Teams     []Team
}

func (q *Queries) FindUsersByTeamID(ctx context.Context, teamID resource.ID) ([]FindUsersByTeamIDRow, error) {
	rows, err := q.db.Query(ctx, findUsersByTeamID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindUsersByTeamIDRow
	for rows.Next() {
		var i FindUsersByTeamIDRow
		if err := rows.Scan(
			&i.UserID,
			&i.Username,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.SiteAdmin,
			&i.Teams,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertUser = `-- name: InsertUser :exec
INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username
) VALUES (
    $1,
    $2,
    $3,
    $4
)
`

type InsertUserParams struct {
	ID        resource.ID
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	Username  pgtype.Text
}

func (q *Queries) InsertUser(ctx context.Context, arg InsertUserParams) error {
	_, err := q.db.Exec(ctx, insertUser,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Username,
	)
	return err
}

const resetUserSiteAdmins = `-- name: ResetUserSiteAdmins :many
UPDATE users
SET site_admin = false
WHERE site_admin = true
RETURNING username
`

func (q *Queries) ResetUserSiteAdmins(ctx context.Context) ([]pgtype.Text, error) {
	rows, err := q.db.Query(ctx, resetUserSiteAdmins)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.Text
	for rows.Next() {
		var username pgtype.Text
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		items = append(items, username)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUserSiteAdmins = `-- name: UpdateUserSiteAdmins :many
UPDATE users
SET site_admin = true
WHERE username = ANY($1::text[])
RETURNING username
`

func (q *Queries) UpdateUserSiteAdmins(ctx context.Context, usernames []pgtype.Text) ([]pgtype.Text, error) {
	rows, err := q.db.Query(ctx, updateUserSiteAdmins, usernames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.Text
	for rows.Next() {
		var username pgtype.Text
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		items = append(items, username)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
