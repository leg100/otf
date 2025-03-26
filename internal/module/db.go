package module

import (
	"context"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is the registry database on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) createModule(ctx context.Context, mod *Module) error {
	_, err := db.Exec(ctx, `
INSERT INTO modules (
    module_id,
    created_at,
    updated_at,
    name,
    provider,
    status,
    organization_name
) VALUES (
	@module_id,
	@created_at,
	@updated_at,
	@name,
	@provider,
	@status,
	@organization_name
)
`, pgx.NamedArgs{
		"module_id":         mod.ID,
		"created_at":        sql.Timestamptz(mod.CreatedAt),
		"updated_at":        sql.Timestamptz(mod.UpdatedAt),
		"name":              sql.String(mod.Name),
		"provider":          sql.String(mod.Provider),
		"status":            sql.String(string(mod.Status)),
		"organization_name": mod.Organization,
	})
	return err
}

func (db *pgdb) updateModuleStatus(ctx context.Context, moduleID resource.TfeID, status ModuleStatus) error {
	_, err := db.Exec(ctx, `
UPDATE modules
SET status = $1
WHERE module_id = $2
RETURNING module_id
`, status, moduleID)
	return err
}

func (db *pgdb) listModules(ctx context.Context, opts ListModulesOptions) ([]*Module, error) {
	rows := db.Query(ctx, `
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
	(r.*)::"repo_connections" AS connection,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.organization_name = $1
`, opts.Organization)
	return sql.CollectRows(rows, db.scanModule)
}

func (db *pgdb) getModule(ctx context.Context, opts GetModuleOptions) (*Module, error) {
	rows := db.Query(ctx, `
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
	(r.*)::"repo_connections" AS connection,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.organization_name = $1
AND   m.name = $2
AND   m.provider = $3
`, opts.Organization, opts.Name, opts.Provider)
	return sql.CollectOneRow(rows, db.scanModule)
}

func (db *pgdb) getModuleByID(ctx context.Context, id resource.TfeID) (*Module, error) {
	rows := db.Query(ctx, `
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
	(r.*)::"repo_connections" AS connection,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.module_id = $1
`, id)
	return sql.CollectOneRow(rows, db.scanModule)
}

func (db *pgdb) getModuleByConnection(ctx context.Context, vcsProviderID resource.TfeID, repoPath string) (*Module, error) {
	rows := db.Query(ctx, `
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
	(r.*)::"repo_connections" AS connection,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
JOIN repo_connections r USING (module_id)
WHERE r.vcs_provider_id = $1
AND   r.repo_path = $2
`, vcsProviderID, repoPath)
	return sql.CollectOneRow(rows, db.scanModule)
}

func (db *pgdb) delete(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM modules
WHERE module_id = $1
`, id)
	return err
}

func (db *pgdb) createModuleVersion(ctx context.Context, version *ModuleVersion) error {
	_, err := db.Exec(ctx, `
INSERT INTO module_versions (
    module_version_id,
    version,
    created_at,
    updated_at,
    module_id,
    status
) VALUES (
	@id,
	@version,
	@created_at,
	@updated_at,
	@module_id,
	@status
)`, pgx.NamedArgs{
		"id":         version.ID,
		"version":    version.Version,
		"created_at": version.CreatedAt,
		"updated_at": version.UpdatedAt,
		"module_id":  version.ModuleID,
		"status":     version.Status,
	})
	return err
}

func (db *pgdb) updateModuleVersionStatus(ctx context.Context, opts UpdateModuleVersionStatusOptions) error {
	_, err := db.Exec(ctx, `
UPDATE module_versions
SET
    status = $1,
    status_error = $2
WHERE module_version_id = $3
`, opts.Status, opts.Error, opts.ID)
	return err
}

func (db *pgdb) getModuleByVersionID(ctx context.Context, versionID resource.TfeID) (*Module, error) {
	rows := db.Query(ctx, `
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
	(r.*)::"repo_connections" AS connection,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
JOIN module_versions mv USING (module_id)
LEFT JOIN repo_connections r USING (module_id)
WHERE mv.module_version_id = $1
`, versionID)
	return sql.CollectOneRow(rows, db.scanModule)
}

func (db *pgdb) deleteModuleVersion(ctx context.Context, versionID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM module_versions
WHERE module_version_id = $1
`, versionID)
	return err
}

func (db *pgdb) saveTarball(ctx context.Context, versionID resource.TfeID, tarball []byte) error {
	_, err := db.Exec(ctx, `
INSERT INTO module_tarballs (
    tarball,
    module_version_id
) VALUES (
    $1,
    $2
)`, tarball, versionID)
	return err
}

func (db *pgdb) getTarball(ctx context.Context, versionID resource.TfeID) ([]byte, error) {
	rows := db.Query(ctx, `
SELECT tarball
FROM module_tarballs
WHERE module_version_id = $1
`, versionID)
	return sql.CollectOneRow(rows, pgx.RowTo[[]byte])
}

func (db *pgdb) scanModule(row pgx.CollectableRow) (*Module, error) {
	mod, err := pgx.RowToAddrOfStructByName[Module](row)
	if err != nil {
		return nil, err
	}
	// versions are always maintained in descending order
	//
	// TODO: this invariant should be part of a constructor
	slices.SortFunc(mod.Versions, func(a, b ModuleVersion) int {
		return semver.Compare(a.Version, b.Version) * -1
	})
	return mod, nil
}
