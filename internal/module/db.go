package module

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type (
	// pgdb is the registry database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// moduleRow is a row from a database query for modules.
	moduleRow struct {
		ModuleID         pgtype.Text
		CreatedAt        pgtype.Timestamptz
		UpdatedAt        pgtype.Timestamptz
		Name             pgtype.Text
		Provider         pgtype.Text
		Status           pgtype.Text
		OrganizationName pgtype.Text
		VCSProviderID    pgtype.Text
		RepoPath         pgtype.Text
		ModuleVersions   []sqlc.ModuleVersion
	}
)

func (db *pgdb) createModule(ctx context.Context, mod *Module) error {
	err := db.Querier(ctx).InsertModule(ctx, sqlc.InsertModuleParams{
		ID:               sql.String(mod.ID),
		CreatedAt:        sql.Timestamptz(mod.CreatedAt),
		UpdatedAt:        sql.Timestamptz(mod.UpdatedAt),
		Name:             sql.String(mod.Name),
		Provider:         sql.String(mod.Provider),
		Status:           sql.String(string(mod.Status)),
		OrganizationName: sql.String(mod.Organization),
	})
	return sql.Error(err)
}

func (db *pgdb) updateModuleStatus(ctx context.Context, moduleID string, status ModuleStatus) error {
	_, err := db.Querier(ctx).UpdateModuleStatusByID(ctx, sqlc.UpdateModuleStatusByIDParams{
		Status:   sql.String(string(status)),
		ModuleID: sql.String(moduleID),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) listModules(ctx context.Context, opts ListModulesOptions) ([]*Module, error) {
	rows, err := db.Querier(ctx).ListModulesByOrganization(ctx, sql.String(opts.Organization))
	if err != nil {
		return nil, err
	}

	modules := make([]*Module, len(rows))
	for i, r := range rows {
		modules[i] = moduleRow(r).toModule()
	}
	return modules, nil
}

func (db *pgdb) getModule(ctx context.Context, opts GetModuleOptions) (*Module, error) {
	row, err := db.Querier(ctx).FindModuleByName(ctx, sqlc.FindModuleByNameParams{
		Name:             sql.String(opts.Name),
		Provider:         sql.String(opts.Provider),
		OrganizationName: sql.String(opts.Organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) getModuleByID(ctx context.Context, id string) (*Module, error) {
	row, err := db.Querier(ctx).FindModuleByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) getModuleByConnection(ctx context.Context, vcsProviderID, repoPath string) (*Module, error) {
	row, err := db.Querier(ctx).FindModuleByConnection(ctx, sqlc.FindModuleByConnectionParams{
		VCSProviderID: sql.String(vcsProviderID),
		RepoPath:      sql.String(repoPath),
	})
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	_, err := db.Querier(ctx).DeleteModuleByID(ctx, sql.String(id))
	return sql.Error(err)
}

func (db *pgdb) createModuleVersion(ctx context.Context, version *ModuleVersion) error {
	_, err := db.Querier(ctx).InsertModuleVersion(ctx, sqlc.InsertModuleVersionParams{
		ModuleVersionID: sql.String(version.ID),
		Version:         sql.String(version.Version),
		CreatedAt:       sql.Timestamptz(version.CreatedAt),
		UpdatedAt:       sql.Timestamptz(version.UpdatedAt),
		ModuleID:        sql.String(version.ModuleID),
		Status:          sql.String(string(version.Status)),
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *pgdb) updateModuleVersionStatus(ctx context.Context, opts UpdateModuleVersionStatusOptions) error {
	_, err := db.Querier(ctx).UpdateModuleVersionStatusByID(ctx, sqlc.UpdateModuleVersionStatusByIDParams{
		ModuleVersionID: sql.String(opts.ID),
		Status:          sql.String(string(opts.Status)),
		StatusError:     sql.String(opts.Error),
	})
	return sql.Error(err)
}

func (db *pgdb) getModuleByVersionID(ctx context.Context, versionID string) (*Module, error) {
	row, err := db.Querier(ctx).FindModuleByModuleVersionID(ctx, sql.String(versionID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return moduleRow(row).toModule(), nil
}

func (db *pgdb) deleteModuleVersion(ctx context.Context, versionID string) error {
	_, err := db.Querier(ctx).DeleteModuleVersionByID(ctx, sql.String(versionID))
	return sql.Error(err)
}

func (db *pgdb) saveTarball(ctx context.Context, versionID string, tarball []byte) error {
	_, err := db.Querier(ctx).InsertModuleTarball(ctx, sqlc.InsertModuleTarballParams{
		Tarball:         tarball,
		ModuleVersionID: sql.String(versionID),
	})
	return sql.Error(err)
}

func (db *pgdb) getTarball(ctx context.Context, versionID string) ([]byte, error) {
	tarball, err := db.Querier(ctx).FindModuleTarball(ctx, sql.String(versionID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return tarball, nil
}

// toModule converts a database row into a module
func (row moduleRow) toModule() *Module {
	module := &Module{
		ID:           row.ModuleID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		UpdatedAt:    row.UpdatedAt.Time.UTC(),
		Name:         row.Name.String,
		Provider:     row.Provider.String,
		Status:       ModuleStatus(row.Status.String),
		Organization: row.OrganizationName.String,
	}
	if row.VCSProviderID.Valid && row.RepoPath.Valid {
		module.Connection = &connections.Connection{
			VCSProviderID: row.VCSProviderID.String,
			Repo:          row.RepoPath.String,
		}
	}
	// versions are always maintained in descending order
	//
	// TODO: this invariant should be part of a constructor
	sort.Sort(byVersion(row.ModuleVersions))
	for i := len(row.ModuleVersions) - 1; i >= 0; i-- {
		module.Versions = append(module.Versions, ModuleVersion{
			ID:          row.ModuleVersions[i].ModuleVersionID.String,
			Version:     row.ModuleVersions[i].Version.String,
			CreatedAt:   row.ModuleVersions[i].CreatedAt.Time.UTC(),
			UpdatedAt:   row.ModuleVersions[i].UpdatedAt.Time.UTC(),
			ModuleID:    row.ModuleVersions[i].ModuleID.String,
			Status:      ModuleVersionStatus(row.ModuleVersions[i].Status.String),
			StatusError: row.ModuleVersions[i].StatusError.String,
		})
	}
	return module
}

type byVersion []sqlc.ModuleVersion

func (v byVersion) Len() int      { return len(v) }
func (v byVersion) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v byVersion) Less(i, j int) bool {
	return semver.Compare(v[i].Version.String, v[j].Version.String) < 0
}
