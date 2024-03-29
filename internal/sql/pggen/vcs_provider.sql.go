// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const insertVCSProviderSQL = `INSERT INTO vcs_providers (
    vcs_provider_id,
    created_at,
    name,
    vcs_kind,
    token,
    github_app_id,
    organization_name
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
);`

type InsertVCSProviderParams struct {
	VCSProviderID    pgtype.Text
	CreatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	VCSKind          pgtype.Text
	Token            pgtype.Text
	GithubAppID      pgtype.Int8
	OrganizationName pgtype.Text
}

// InsertVCSProvider implements Querier.InsertVCSProvider.
func (q *DBQuerier) InsertVCSProvider(ctx context.Context, params InsertVCSProviderParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertVCSProvider")
	cmdTag, err := q.conn.Exec(ctx, insertVCSProviderSQL, params.VCSProviderID, params.CreatedAt, params.Name, params.VCSKind, params.Token, params.GithubAppID, params.OrganizationName)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertVCSProvider: %w", err)
	}
	return cmdTag, err
}

// InsertVCSProviderBatch implements Querier.InsertVCSProviderBatch.
func (q *DBQuerier) InsertVCSProviderBatch(batch genericBatch, params InsertVCSProviderParams) {
	batch.Queue(insertVCSProviderSQL, params.VCSProviderID, params.CreatedAt, params.Name, params.VCSKind, params.Token, params.GithubAppID, params.OrganizationName)
}

// InsertVCSProviderScan implements Querier.InsertVCSProviderScan.
func (q *DBQuerier) InsertVCSProviderScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertVCSProviderBatch: %w", err)
	}
	return cmdTag, err
}

const findVCSProvidersByOrganizationSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.organization_name = $1
;`

type FindVCSProvidersByOrganizationRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProvidersByOrganization implements Querier.FindVCSProvidersByOrganization.
func (q *DBQuerier) FindVCSProvidersByOrganization(ctx context.Context, organizationName pgtype.Text) ([]FindVCSProvidersByOrganizationRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProvidersByOrganization")
	rows, err := q.conn.Query(ctx, findVCSProvidersByOrganizationSQL, organizationName)
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProvidersByOrganization: %w", err)
	}
	defer rows.Close()
	items := []FindVCSProvidersByOrganizationRow{}
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	for rows.Next() {
		var item FindVCSProvidersByOrganizationRow
		if err := rows.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
			return nil, fmt.Errorf("scan FindVCSProvidersByOrganization row: %w", err)
		}
		if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
			return nil, fmt.Errorf("assign FindVCSProvidersByOrganization row: %w", err)
		}
		if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
			return nil, fmt.Errorf("assign FindVCSProvidersByOrganization row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindVCSProvidersByOrganization rows: %w", err)
	}
	return items, err
}

// FindVCSProvidersByOrganizationBatch implements Querier.FindVCSProvidersByOrganizationBatch.
func (q *DBQuerier) FindVCSProvidersByOrganizationBatch(batch genericBatch, organizationName pgtype.Text) {
	batch.Queue(findVCSProvidersByOrganizationSQL, organizationName)
}

// FindVCSProvidersByOrganizationScan implements Querier.FindVCSProvidersByOrganizationScan.
func (q *DBQuerier) FindVCSProvidersByOrganizationScan(results pgx.BatchResults) ([]FindVCSProvidersByOrganizationRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProvidersByOrganizationBatch: %w", err)
	}
	defer rows.Close()
	items := []FindVCSProvidersByOrganizationRow{}
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	for rows.Next() {
		var item FindVCSProvidersByOrganizationRow
		if err := rows.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
			return nil, fmt.Errorf("scan FindVCSProvidersByOrganizationBatch row: %w", err)
		}
		if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
			return nil, fmt.Errorf("assign FindVCSProvidersByOrganization row: %w", err)
		}
		if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
			return nil, fmt.Errorf("assign FindVCSProvidersByOrganization row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindVCSProvidersByOrganizationBatch rows: %w", err)
	}
	return items, err
}

const findVCSProvidersSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
;`

type FindVCSProvidersRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProviders implements Querier.FindVCSProviders.
func (q *DBQuerier) FindVCSProviders(ctx context.Context) ([]FindVCSProvidersRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProviders")
	rows, err := q.conn.Query(ctx, findVCSProvidersSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProviders: %w", err)
	}
	defer rows.Close()
	items := []FindVCSProvidersRow{}
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	for rows.Next() {
		var item FindVCSProvidersRow
		if err := rows.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
			return nil, fmt.Errorf("scan FindVCSProviders row: %w", err)
		}
		if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
			return nil, fmt.Errorf("assign FindVCSProviders row: %w", err)
		}
		if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
			return nil, fmt.Errorf("assign FindVCSProviders row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindVCSProviders rows: %w", err)
	}
	return items, err
}

// FindVCSProvidersBatch implements Querier.FindVCSProvidersBatch.
func (q *DBQuerier) FindVCSProvidersBatch(batch genericBatch) {
	batch.Queue(findVCSProvidersSQL)
}

// FindVCSProvidersScan implements Querier.FindVCSProvidersScan.
func (q *DBQuerier) FindVCSProvidersScan(results pgx.BatchResults) ([]FindVCSProvidersRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProvidersBatch: %w", err)
	}
	defer rows.Close()
	items := []FindVCSProvidersRow{}
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	for rows.Next() {
		var item FindVCSProvidersRow
		if err := rows.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
			return nil, fmt.Errorf("scan FindVCSProvidersBatch row: %w", err)
		}
		if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
			return nil, fmt.Errorf("assign FindVCSProviders row: %w", err)
		}
		if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
			return nil, fmt.Errorf("assign FindVCSProviders row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindVCSProvidersBatch rows: %w", err)
	}
	return items, err
}

const findVCSProvidersByGithubAppInstallIDSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE gi.install_id = $1
;`

type FindVCSProvidersByGithubAppInstallIDRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProvidersByGithubAppInstallID implements Querier.FindVCSProvidersByGithubAppInstallID.
func (q *DBQuerier) FindVCSProvidersByGithubAppInstallID(ctx context.Context, installID pgtype.Int8) ([]FindVCSProvidersByGithubAppInstallIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProvidersByGithubAppInstallID")
	rows, err := q.conn.Query(ctx, findVCSProvidersByGithubAppInstallIDSQL, installID)
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProvidersByGithubAppInstallID: %w", err)
	}
	defer rows.Close()
	items := []FindVCSProvidersByGithubAppInstallIDRow{}
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	for rows.Next() {
		var item FindVCSProvidersByGithubAppInstallIDRow
		if err := rows.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
			return nil, fmt.Errorf("scan FindVCSProvidersByGithubAppInstallID row: %w", err)
		}
		if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
			return nil, fmt.Errorf("assign FindVCSProvidersByGithubAppInstallID row: %w", err)
		}
		if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
			return nil, fmt.Errorf("assign FindVCSProvidersByGithubAppInstallID row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindVCSProvidersByGithubAppInstallID rows: %w", err)
	}
	return items, err
}

// FindVCSProvidersByGithubAppInstallIDBatch implements Querier.FindVCSProvidersByGithubAppInstallIDBatch.
func (q *DBQuerier) FindVCSProvidersByGithubAppInstallIDBatch(batch genericBatch, installID pgtype.Int8) {
	batch.Queue(findVCSProvidersByGithubAppInstallIDSQL, installID)
}

// FindVCSProvidersByGithubAppInstallIDScan implements Querier.FindVCSProvidersByGithubAppInstallIDScan.
func (q *DBQuerier) FindVCSProvidersByGithubAppInstallIDScan(results pgx.BatchResults) ([]FindVCSProvidersByGithubAppInstallIDRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProvidersByGithubAppInstallIDBatch: %w", err)
	}
	defer rows.Close()
	items := []FindVCSProvidersByGithubAppInstallIDRow{}
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	for rows.Next() {
		var item FindVCSProvidersByGithubAppInstallIDRow
		if err := rows.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
			return nil, fmt.Errorf("scan FindVCSProvidersByGithubAppInstallIDBatch row: %w", err)
		}
		if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
			return nil, fmt.Errorf("assign FindVCSProvidersByGithubAppInstallID row: %w", err)
		}
		if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
			return nil, fmt.Errorf("assign FindVCSProvidersByGithubAppInstallID row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindVCSProvidersByGithubAppInstallIDBatch rows: %w", err)
	}
	return items, err
}

const findVCSProviderSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.vcs_provider_id = $1
;`

type FindVCSProviderRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProvider implements Querier.FindVCSProvider.
func (q *DBQuerier) FindVCSProvider(ctx context.Context, vcsProviderID pgtype.Text) (FindVCSProviderRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProvider")
	row := q.conn.QueryRow(ctx, findVCSProviderSQL, vcsProviderID)
	var item FindVCSProviderRow
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	if err := row.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
		return item, fmt.Errorf("query FindVCSProvider: %w", err)
	}
	if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
		return item, fmt.Errorf("assign FindVCSProvider row: %w", err)
	}
	if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
		return item, fmt.Errorf("assign FindVCSProvider row: %w", err)
	}
	return item, nil
}

// FindVCSProviderBatch implements Querier.FindVCSProviderBatch.
func (q *DBQuerier) FindVCSProviderBatch(batch genericBatch, vcsProviderID pgtype.Text) {
	batch.Queue(findVCSProviderSQL, vcsProviderID)
}

// FindVCSProviderScan implements Querier.FindVCSProviderScan.
func (q *DBQuerier) FindVCSProviderScan(results pgx.BatchResults) (FindVCSProviderRow, error) {
	row := results.QueryRow()
	var item FindVCSProviderRow
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	if err := row.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
		return item, fmt.Errorf("scan FindVCSProviderBatch row: %w", err)
	}
	if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
		return item, fmt.Errorf("assign FindVCSProvider row: %w", err)
	}
	if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
		return item, fmt.Errorf("assign FindVCSProvider row: %w", err)
	}
	return item, nil
}

const findVCSProviderForUpdateSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.vcs_provider_id = $1
FOR UPDATE OF v
;`

type FindVCSProviderForUpdateRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProviderForUpdate implements Querier.FindVCSProviderForUpdate.
func (q *DBQuerier) FindVCSProviderForUpdate(ctx context.Context, vcsProviderID pgtype.Text) (FindVCSProviderForUpdateRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProviderForUpdate")
	row := q.conn.QueryRow(ctx, findVCSProviderForUpdateSQL, vcsProviderID)
	var item FindVCSProviderForUpdateRow
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	if err := row.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
		return item, fmt.Errorf("query FindVCSProviderForUpdate: %w", err)
	}
	if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
		return item, fmt.Errorf("assign FindVCSProviderForUpdate row: %w", err)
	}
	if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
		return item, fmt.Errorf("assign FindVCSProviderForUpdate row: %w", err)
	}
	return item, nil
}

// FindVCSProviderForUpdateBatch implements Querier.FindVCSProviderForUpdateBatch.
func (q *DBQuerier) FindVCSProviderForUpdateBatch(batch genericBatch, vcsProviderID pgtype.Text) {
	batch.Queue(findVCSProviderForUpdateSQL, vcsProviderID)
}

// FindVCSProviderForUpdateScan implements Querier.FindVCSProviderForUpdateScan.
func (q *DBQuerier) FindVCSProviderForUpdateScan(results pgx.BatchResults) (FindVCSProviderForUpdateRow, error) {
	row := results.QueryRow()
	var item FindVCSProviderForUpdateRow
	githubAppRow := q.types.newGithubApps()
	githubAppInstallRow := q.types.newGithubAppInstalls()
	if err := row.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID, githubAppRow, githubAppInstallRow); err != nil {
		return item, fmt.Errorf("scan FindVCSProviderForUpdateBatch row: %w", err)
	}
	if err := githubAppRow.AssignTo(&item.GithubApp); err != nil {
		return item, fmt.Errorf("assign FindVCSProviderForUpdate row: %w", err)
	}
	if err := githubAppInstallRow.AssignTo(&item.GithubAppInstall); err != nil {
		return item, fmt.Errorf("assign FindVCSProviderForUpdate row: %w", err)
	}
	return item, nil
}

const updateVCSProviderSQL = `UPDATE vcs_providers
SET name = $1, token = $2
WHERE vcs_provider_id = $3
RETURNING *
;`

type UpdateVCSProviderParams struct {
	Name          pgtype.Text
	Token         pgtype.Text
	VCSProviderID pgtype.Text
}

type UpdateVCSProviderRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
}

// UpdateVCSProvider implements Querier.UpdateVCSProvider.
func (q *DBQuerier) UpdateVCSProvider(ctx context.Context, params UpdateVCSProviderParams) (UpdateVCSProviderRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateVCSProvider")
	row := q.conn.QueryRow(ctx, updateVCSProviderSQL, params.Name, params.Token, params.VCSProviderID)
	var item UpdateVCSProviderRow
	if err := row.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID); err != nil {
		return item, fmt.Errorf("query UpdateVCSProvider: %w", err)
	}
	return item, nil
}

// UpdateVCSProviderBatch implements Querier.UpdateVCSProviderBatch.
func (q *DBQuerier) UpdateVCSProviderBatch(batch genericBatch, params UpdateVCSProviderParams) {
	batch.Queue(updateVCSProviderSQL, params.Name, params.Token, params.VCSProviderID)
}

// UpdateVCSProviderScan implements Querier.UpdateVCSProviderScan.
func (q *DBQuerier) UpdateVCSProviderScan(results pgx.BatchResults) (UpdateVCSProviderRow, error) {
	row := results.QueryRow()
	var item UpdateVCSProviderRow
	if err := row.Scan(&item.VCSProviderID, &item.Token, &item.CreatedAt, &item.Name, &item.VCSKind, &item.OrganizationName, &item.GithubAppID); err != nil {
		return item, fmt.Errorf("scan UpdateVCSProviderBatch row: %w", err)
	}
	return item, nil
}

const deleteVCSProviderByIDSQL = `DELETE
FROM vcs_providers
WHERE vcs_provider_id = $1
RETURNING vcs_provider_id
;`

// DeleteVCSProviderByID implements Querier.DeleteVCSProviderByID.
func (q *DBQuerier) DeleteVCSProviderByID(ctx context.Context, vcsProviderID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteVCSProviderByID")
	row := q.conn.QueryRow(ctx, deleteVCSProviderByIDSQL, vcsProviderID)
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query DeleteVCSProviderByID: %w", err)
	}
	return item, nil
}

// DeleteVCSProviderByIDBatch implements Querier.DeleteVCSProviderByIDBatch.
func (q *DBQuerier) DeleteVCSProviderByIDBatch(batch genericBatch, vcsProviderID pgtype.Text) {
	batch.Queue(deleteVCSProviderByIDSQL, vcsProviderID)
}

// DeleteVCSProviderByIDScan implements Querier.DeleteVCSProviderByIDScan.
func (q *DBQuerier) DeleteVCSProviderByIDScan(results pgx.BatchResults) (pgtype.Text, error) {
	row := results.QueryRow()
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan DeleteVCSProviderByIDBatch row: %w", err)
	}
	return item, nil
}
