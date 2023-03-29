// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const insertUserSQL = `INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username
) VALUES (
    $1,
    $2,
    $3,
    $4
);`

type InsertUserParams struct {
	ID        pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	Username  pgtype.Text
}

// InsertUser implements Querier.InsertUser.
func (q *DBQuerier) InsertUser(ctx context.Context, params InsertUserParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertUser")
	cmdTag, err := q.conn.Exec(ctx, insertUserSQL, params.ID, params.CreatedAt, params.UpdatedAt, params.Username)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertUser: %w", err)
	}
	return cmdTag, err
}

// InsertUserBatch implements Querier.InsertUserBatch.
func (q *DBQuerier) InsertUserBatch(batch genericBatch, params InsertUserParams) {
	batch.Queue(insertUserSQL, params.ID, params.CreatedAt, params.UpdatedAt, params.Username)
}

// InsertUserScan implements Querier.InsertUserScan.
func (q *DBQuerier) InsertUserScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertUserBatch: %w", err)
	}
	return cmdTag, err
}

const findUsersSQL = `SELECT u.*,
    array_remove(array_agg(o.name), NULL) AS organizations,
    array_remove(array_agg(teams), NULL) AS teams
FROM users u
LEFT JOIN (organization_memberships om JOIN organizations o ON om.organization_name = o.name) ON u.username = om.username
LEFT JOIN (team_memberships tm JOIN teams USING (team_id)) ON u.username = tm.username
GROUP BY u.user_id
;`

type FindUsersRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []Teams            `json:"teams"`
}

// FindUsers implements Querier.FindUsers.
func (q *DBQuerier) FindUsers(ctx context.Context) ([]FindUsersRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUsers")
	rows, err := q.conn.Query(ctx, findUsersSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindUsers: %w", err)
	}
	defer rows.Close()
	items := []FindUsersRow{}
	teamsArray := q.types.newTeamsArray()
	for rows.Next() {
		var item FindUsersRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsers row: %w", err)
		}
		if err := teamsArray.AssignTo(&item.Teams); err != nil {
			return nil, fmt.Errorf("assign FindUsers row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsers rows: %w", err)
	}
	return items, err
}

// FindUsersBatch implements Querier.FindUsersBatch.
func (q *DBQuerier) FindUsersBatch(batch genericBatch) {
	batch.Queue(findUsersSQL)
}

// FindUsersScan implements Querier.FindUsersScan.
func (q *DBQuerier) FindUsersScan(results pgx.BatchResults) ([]FindUsersRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindUsersBatch: %w", err)
	}
	defer rows.Close()
	items := []FindUsersRow{}
	teamsArray := q.types.newTeamsArray()
	for rows.Next() {
		var item FindUsersRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsersBatch row: %w", err)
		}
		if err := teamsArray.AssignTo(&item.Teams); err != nil {
			return nil, fmt.Errorf("assign FindUsers row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsersBatch rows: %w", err)
	}
	return items, err
}

const findUsersByOrganizationSQL = `SELECT u.*,
    (
        SELECT array_remove(array_agg(o.name), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.username = u.username
    ) AS organizations,
    array_remove(array_agg(teams), NULL) AS teams
FROM users u
JOIN (organization_memberships om JOIN organizations o ON om.organization_name = o.name) ON u.username = om.username
LEFT JOIN (team_memberships tm JOIN teams USING (team_id)) ON u.username = tm.username
WHERE o.name = $1
GROUP BY u.user_id
;`

type FindUsersByOrganizationRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []Teams            `json:"teams"`
}

// FindUsersByOrganization implements Querier.FindUsersByOrganization.
func (q *DBQuerier) FindUsersByOrganization(ctx context.Context, organizationName pgtype.Text) ([]FindUsersByOrganizationRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUsersByOrganization")
	rows, err := q.conn.Query(ctx, findUsersByOrganizationSQL, organizationName)
	if err != nil {
		return nil, fmt.Errorf("query FindUsersByOrganization: %w", err)
	}
	defer rows.Close()
	items := []FindUsersByOrganizationRow{}
	teamsArray := q.types.newTeamsArray()
	for rows.Next() {
		var item FindUsersByOrganizationRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsersByOrganization row: %w", err)
		}
		if err := teamsArray.AssignTo(&item.Teams); err != nil {
			return nil, fmt.Errorf("assign FindUsersByOrganization row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsersByOrganization rows: %w", err)
	}
	return items, err
}

// FindUsersByOrganizationBatch implements Querier.FindUsersByOrganizationBatch.
func (q *DBQuerier) FindUsersByOrganizationBatch(batch genericBatch, organizationName pgtype.Text) {
	batch.Queue(findUsersByOrganizationSQL, organizationName)
}

// FindUsersByOrganizationScan implements Querier.FindUsersByOrganizationScan.
func (q *DBQuerier) FindUsersByOrganizationScan(results pgx.BatchResults) ([]FindUsersByOrganizationRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindUsersByOrganizationBatch: %w", err)
	}
	defer rows.Close()
	items := []FindUsersByOrganizationRow{}
	teamsArray := q.types.newTeamsArray()
	for rows.Next() {
		var item FindUsersByOrganizationRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsersByOrganizationBatch row: %w", err)
		}
		if err := teamsArray.AssignTo(&item.Teams); err != nil {
			return nil, fmt.Errorf("assign FindUsersByOrganization row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsersByOrganizationBatch rows: %w", err)
	}
	return items, err
}

const findUsersByTeamSQL = `SELECT
    u.*,
    array_remove(array_agg(o.name), NULL) AS organizations,
    array_remove(array_agg(t), NULL) AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
JOIN organizations o ON o.name = t.organization_name
WHERE o.name = $1
AND   t.name = $2
GROUP BY u.user_id
;`

type FindUsersByTeamRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []Teams            `json:"teams"`
}

// FindUsersByTeam implements Querier.FindUsersByTeam.
func (q *DBQuerier) FindUsersByTeam(ctx context.Context, organizationName pgtype.Text, teamName pgtype.Text) ([]FindUsersByTeamRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUsersByTeam")
	rows, err := q.conn.Query(ctx, findUsersByTeamSQL, organizationName, teamName)
	if err != nil {
		return nil, fmt.Errorf("query FindUsersByTeam: %w", err)
	}
	defer rows.Close()
	items := []FindUsersByTeamRow{}
	teamsArray := q.types.newTeamsArray()
	for rows.Next() {
		var item FindUsersByTeamRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsersByTeam row: %w", err)
		}
		if err := teamsArray.AssignTo(&item.Teams); err != nil {
			return nil, fmt.Errorf("assign FindUsersByTeam row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsersByTeam rows: %w", err)
	}
	return items, err
}

// FindUsersByTeamBatch implements Querier.FindUsersByTeamBatch.
func (q *DBQuerier) FindUsersByTeamBatch(batch genericBatch, organizationName pgtype.Text, teamName pgtype.Text) {
	batch.Queue(findUsersByTeamSQL, organizationName, teamName)
}

// FindUsersByTeamScan implements Querier.FindUsersByTeamScan.
func (q *DBQuerier) FindUsersByTeamScan(results pgx.BatchResults) ([]FindUsersByTeamRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindUsersByTeamBatch: %w", err)
	}
	defer rows.Close()
	items := []FindUsersByTeamRow{}
	teamsArray := q.types.newTeamsArray()
	for rows.Next() {
		var item FindUsersByTeamRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsersByTeamBatch row: %w", err)
		}
		if err := teamsArray.AssignTo(&item.Teams); err != nil {
			return nil, fmt.Errorf("assign FindUsersByTeam row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsersByTeamBatch rows: %w", err)
	}
	return items, err
}

const findUsersByTeamIDSQL = `SELECT
    u.*,
    (
        SELECT array_agg(o.name)
        FROM organizations o
        JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.username = u.username
    ) AS organizations,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.team_id = $1
GROUP BY u.user_id
;`

type FindUsersByTeamIDRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []Teams            `json:"teams"`
}

// FindUsersByTeamID implements Querier.FindUsersByTeamID.
func (q *DBQuerier) FindUsersByTeamID(ctx context.Context, teamID pgtype.Text) ([]FindUsersByTeamIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUsersByTeamID")
	rows, err := q.conn.Query(ctx, findUsersByTeamIDSQL, teamID)
	if err != nil {
		return nil, fmt.Errorf("query FindUsersByTeamID: %w", err)
	}
	defer rows.Close()
	items := []FindUsersByTeamIDRow{}
	teamsArray := q.types.newTeamsArray()
	for rows.Next() {
		var item FindUsersByTeamIDRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsersByTeamID row: %w", err)
		}
		if err := teamsArray.AssignTo(&item.Teams); err != nil {
			return nil, fmt.Errorf("assign FindUsersByTeamID row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsersByTeamID rows: %w", err)
	}
	return items, err
}

// FindUsersByTeamIDBatch implements Querier.FindUsersByTeamIDBatch.
func (q *DBQuerier) FindUsersByTeamIDBatch(batch genericBatch, teamID pgtype.Text) {
	batch.Queue(findUsersByTeamIDSQL, teamID)
}

// FindUsersByTeamIDScan implements Querier.FindUsersByTeamIDScan.
func (q *DBQuerier) FindUsersByTeamIDScan(results pgx.BatchResults) ([]FindUsersByTeamIDRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindUsersByTeamIDBatch: %w", err)
	}
	defer rows.Close()
	items := []FindUsersByTeamIDRow{}
	teamsArray := q.types.newTeamsArray()
	for rows.Next() {
		var item FindUsersByTeamIDRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsersByTeamIDBatch row: %w", err)
		}
		if err := teamsArray.AssignTo(&item.Teams); err != nil {
			return nil, fmt.Errorf("assign FindUsersByTeamID row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsersByTeamIDBatch rows: %w", err)
	}
	return items, err
}

const findUserByIDSQL = `SELECT u.*,
    (
        SELECT array_remove(array_agg(o.name), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.username = u.username
    ) AS organizations,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM teams t
        LEFT JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
WHERE u.user_id = $1
GROUP BY u.user_id
;`

type FindUserByIDRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []Teams            `json:"teams"`
}

// FindUserByID implements Querier.FindUserByID.
func (q *DBQuerier) FindUserByID(ctx context.Context, userID pgtype.Text) (FindUserByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserByID")
	row := q.conn.QueryRow(ctx, findUserByIDSQL, userID)
	var item FindUserByIDRow
	teamsArray := q.types.newTeamsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
		return item, fmt.Errorf("query FindUserByID: %w", err)
	}
	if err := teamsArray.AssignTo(&item.Teams); err != nil {
		return item, fmt.Errorf("assign FindUserByID row: %w", err)
	}
	return item, nil
}

// FindUserByIDBatch implements Querier.FindUserByIDBatch.
func (q *DBQuerier) FindUserByIDBatch(batch genericBatch, userID pgtype.Text) {
	batch.Queue(findUserByIDSQL, userID)
}

// FindUserByIDScan implements Querier.FindUserByIDScan.
func (q *DBQuerier) FindUserByIDScan(results pgx.BatchResults) (FindUserByIDRow, error) {
	row := results.QueryRow()
	var item FindUserByIDRow
	teamsArray := q.types.newTeamsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
		return item, fmt.Errorf("scan FindUserByIDBatch row: %w", err)
	}
	if err := teamsArray.AssignTo(&item.Teams); err != nil {
		return item, fmt.Errorf("assign FindUserByID row: %w", err)
	}
	return item, nil
}

const findUserByUsernameSQL = `SELECT u.*,
    (
        SELECT array_remove(array_agg(o.name), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.username = u.username
    ) AS organizations,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM teams t
        LEFT JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
WHERE u.username = $1
GROUP BY u.user_id
;`

type FindUserByUsernameRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []Teams            `json:"teams"`
}

// FindUserByUsername implements Querier.FindUserByUsername.
func (q *DBQuerier) FindUserByUsername(ctx context.Context, username pgtype.Text) (FindUserByUsernameRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserByUsername")
	row := q.conn.QueryRow(ctx, findUserByUsernameSQL, username)
	var item FindUserByUsernameRow
	teamsArray := q.types.newTeamsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
		return item, fmt.Errorf("query FindUserByUsername: %w", err)
	}
	if err := teamsArray.AssignTo(&item.Teams); err != nil {
		return item, fmt.Errorf("assign FindUserByUsername row: %w", err)
	}
	return item, nil
}

// FindUserByUsernameBatch implements Querier.FindUserByUsernameBatch.
func (q *DBQuerier) FindUserByUsernameBatch(batch genericBatch, username pgtype.Text) {
	batch.Queue(findUserByUsernameSQL, username)
}

// FindUserByUsernameScan implements Querier.FindUserByUsernameScan.
func (q *DBQuerier) FindUserByUsernameScan(results pgx.BatchResults) (FindUserByUsernameRow, error) {
	row := results.QueryRow()
	var item FindUserByUsernameRow
	teamsArray := q.types.newTeamsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
		return item, fmt.Errorf("scan FindUserByUsernameBatch row: %w", err)
	}
	if err := teamsArray.AssignTo(&item.Teams); err != nil {
		return item, fmt.Errorf("assign FindUserByUsername row: %w", err)
	}
	return item, nil
}

const findUserBySessionTokenSQL = `SELECT u.*,
    (
        SELECT array_agg(o.name)
        FROM organizations o
        JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.username = u.username
    ) AS organizations,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
JOIN sessions s ON u.username = s.username AND s.expiry > current_timestamp
WHERE s.token = $1
GROUP BY u.user_id
;`

type FindUserBySessionTokenRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []Teams            `json:"teams"`
}

// FindUserBySessionToken implements Querier.FindUserBySessionToken.
func (q *DBQuerier) FindUserBySessionToken(ctx context.Context, token pgtype.Text) (FindUserBySessionTokenRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserBySessionToken")
	row := q.conn.QueryRow(ctx, findUserBySessionTokenSQL, token)
	var item FindUserBySessionTokenRow
	teamsArray := q.types.newTeamsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
		return item, fmt.Errorf("query FindUserBySessionToken: %w", err)
	}
	if err := teamsArray.AssignTo(&item.Teams); err != nil {
		return item, fmt.Errorf("assign FindUserBySessionToken row: %w", err)
	}
	return item, nil
}

// FindUserBySessionTokenBatch implements Querier.FindUserBySessionTokenBatch.
func (q *DBQuerier) FindUserBySessionTokenBatch(batch genericBatch, token pgtype.Text) {
	batch.Queue(findUserBySessionTokenSQL, token)
}

// FindUserBySessionTokenScan implements Querier.FindUserBySessionTokenScan.
func (q *DBQuerier) FindUserBySessionTokenScan(results pgx.BatchResults) (FindUserBySessionTokenRow, error) {
	row := results.QueryRow()
	var item FindUserBySessionTokenRow
	teamsArray := q.types.newTeamsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
		return item, fmt.Errorf("scan FindUserBySessionTokenBatch row: %w", err)
	}
	if err := teamsArray.AssignTo(&item.Teams); err != nil {
		return item, fmt.Errorf("assign FindUserBySessionToken row: %w", err)
	}
	return item, nil
}

const findUserByAuthenticationTokenSQL = `SELECT u.*,
    (
        select array_remove(array_agg(o.name), null)
        from organizations o
        left join organization_memberships om ON om.organization_name = o.name
        where om.username = u.username
    ) as organizations,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM teams t
        LEFT JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
LEFT JOIN tokens t ON u.username = t.username
WHERE t.token = $1
GROUP BY u.user_id
;`

type FindUserByAuthenticationTokenRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []Teams            `json:"teams"`
}

// FindUserByAuthenticationToken implements Querier.FindUserByAuthenticationToken.
func (q *DBQuerier) FindUserByAuthenticationToken(ctx context.Context, token pgtype.Text) (FindUserByAuthenticationTokenRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserByAuthenticationToken")
	row := q.conn.QueryRow(ctx, findUserByAuthenticationTokenSQL, token)
	var item FindUserByAuthenticationTokenRow
	teamsArray := q.types.newTeamsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
		return item, fmt.Errorf("query FindUserByAuthenticationToken: %w", err)
	}
	if err := teamsArray.AssignTo(&item.Teams); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationToken row: %w", err)
	}
	return item, nil
}

// FindUserByAuthenticationTokenBatch implements Querier.FindUserByAuthenticationTokenBatch.
func (q *DBQuerier) FindUserByAuthenticationTokenBatch(batch genericBatch, token pgtype.Text) {
	batch.Queue(findUserByAuthenticationTokenSQL, token)
}

// FindUserByAuthenticationTokenScan implements Querier.FindUserByAuthenticationTokenScan.
func (q *DBQuerier) FindUserByAuthenticationTokenScan(results pgx.BatchResults) (FindUserByAuthenticationTokenRow, error) {
	row := results.QueryRow()
	var item FindUserByAuthenticationTokenRow
	teamsArray := q.types.newTeamsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, &item.Organizations, teamsArray); err != nil {
		return item, fmt.Errorf("scan FindUserByAuthenticationTokenBatch row: %w", err)
	}
	if err := teamsArray.AssignTo(&item.Teams); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationToken row: %w", err)
	}
	return item, nil
}

const deleteUserByIDSQL = `DELETE
FROM users
WHERE user_id = $1
RETURNING user_id
;`

// DeleteUserByID implements Querier.DeleteUserByID.
func (q *DBQuerier) DeleteUserByID(ctx context.Context, userID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteUserByID")
	row := q.conn.QueryRow(ctx, deleteUserByIDSQL, userID)
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query DeleteUserByID: %w", err)
	}
	return item, nil
}

// DeleteUserByIDBatch implements Querier.DeleteUserByIDBatch.
func (q *DBQuerier) DeleteUserByIDBatch(batch genericBatch, userID pgtype.Text) {
	batch.Queue(deleteUserByIDSQL, userID)
}

// DeleteUserByIDScan implements Querier.DeleteUserByIDScan.
func (q *DBQuerier) DeleteUserByIDScan(results pgx.BatchResults) (pgtype.Text, error) {
	row := results.QueryRow()
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan DeleteUserByIDBatch row: %w", err)
	}
	return item, nil
}

const deleteUserByUsernameSQL = `DELETE
FROM users
WHERE username = $1
RETURNING user_id
;`

// DeleteUserByUsername implements Querier.DeleteUserByUsername.
func (q *DBQuerier) DeleteUserByUsername(ctx context.Context, username pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteUserByUsername")
	row := q.conn.QueryRow(ctx, deleteUserByUsernameSQL, username)
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query DeleteUserByUsername: %w", err)
	}
	return item, nil
}

// DeleteUserByUsernameBatch implements Querier.DeleteUserByUsernameBatch.
func (q *DBQuerier) DeleteUserByUsernameBatch(batch genericBatch, username pgtype.Text) {
	batch.Queue(deleteUserByUsernameSQL, username)
}

// DeleteUserByUsernameScan implements Querier.DeleteUserByUsernameScan.
func (q *DBQuerier) DeleteUserByUsernameScan(results pgx.BatchResults) (pgtype.Text, error) {
	row := results.QueryRow()
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan DeleteUserByUsernameBatch row: %w", err)
	}
	return item, nil
}
