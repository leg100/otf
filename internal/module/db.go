package module

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

type (
	// pgdb is the registry database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// moduleRow is a row from a database query for modules.
	moduleRow struct {
		ModuleID         resource.ID
		CreatedAt        pgtype.Timestamptz
		UpdatedAt        pgtype.Timestamptz
		Name             pgtype.Text
		Provider         pgtype.Text
		Status           pgtype.Text
		OrganizationName resource.OrganizationName
		VCSProviderID    resource.ID
		RepoPath         pgtype.Text
		ModuleVersions   []ModuleVersionModel
	}
)

func (db *pgdb) createModule(ctx context.Context, mod *Module) error {
	err := q.InsertModule(ctx, db.Conn(ctx), InsertModuleParams{
		ID:               mod.ID,
		CreatedAt:        sql.Timestamptz(mod.CreatedAt),
		UpdatedAt:        sql.Timestamptz(mod.UpdatedAt),
		Name:             sql.String(mod.Name),
		Provider:         sql.String(mod.Provider),
		Status:           sql.String(string(mod.Status)),
		OrganizationName: mod.Organization,
	})
	return sql.Error(err)
}

func (db *pgdb) updateModuleStatus(ctx context.Context, moduleID resource.ID, status ModuleStatus) error {
	_, err := q.UpdateModuleStatusByID(ctx, db.Conn(ctx), UpdateModuleStatusByIDParams{
		Status:   sql.String(string(status)),
		ModuleID: moduleID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) listModules(ctx context.Context, opts ListModulesOptions) ([]*Module, error) {
	rows, err := q.ListModulesByOrganization(ctx, db.Conn(ctx), opts.Organization)
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
	row, err := q.FindModuleByName(ctx, db.Conn(ctx), FindModuleByNameParams{
		Name:             sql.String(opts.Name),
		Provider:         sql.String(opts.Provider),
		OrganizationName: opts.Organization,
	})
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) getModuleByID(ctx context.Context, id resource.ID) (*Module, error) {
	row, err := q.FindModuleByID(ctx, db.Conn(ctx), id)
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) getModuleByConnection(ctx context.Context, vcsProviderID resource.ID, repoPath string) (*Module, error) {
	row, err := q.FindModuleByConnection(ctx, db.Conn(ctx), FindModuleByConnectionParams{
		VCSProviderID: vcsProviderID,
		RepoPath:      sql.String(repoPath),
	})
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) delete(ctx context.Context, id resource.ID) error {
	_, err := q.DeleteModuleByID(ctx, db.Conn(ctx), id)
	return sql.Error(err)
}

func (db *pgdb) createModuleVersion(ctx context.Context, version *ModuleVersion) error {
	_, err := q.InsertModuleVersion(ctx, db.Conn(ctx), InsertModuleVersionParams{
		ModuleVersionID: version.ID,
		Version:         sql.String(version.Version),
		CreatedAt:       sql.Timestamptz(version.CreatedAt),
		UpdatedAt:       sql.Timestamptz(version.UpdatedAt),
		ModuleID:        version.ModuleID,
		Status:          sql.String(string(version.Status)),
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *pgdb) updateModuleVersionStatus(ctx context.Context, opts UpdateModuleVersionStatusOptions) error {
	_, err := q.UpdateModuleVersionStatusByID(ctx, db.Conn(ctx), UpdateModuleVersionStatusByIDParams{
		ModuleVersionID: opts.ID,
		Status:          sql.String(string(opts.Status)),
		StatusError:     sql.String(opts.Error),
	})
	return sql.Error(err)
}

func (db *pgdb) getModuleByVersionID(ctx context.Context, versionID resource.ID) (*Module, error) {
	row, err := q.FindModuleByModuleVersionID(ctx, db.Conn(ctx), versionID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return moduleRow(row).toModule(), nil
}

func (db *pgdb) deleteModuleVersion(ctx context.Context, versionID resource.ID) error {
	_, err := q.DeleteModuleVersionByID(ctx, db.Conn(ctx), versionID)
	return sql.Error(err)
}

func (db *pgdb) saveTarball(ctx context.Context, versionID resource.ID, tarball []byte) error {
	_, err := q.InsertModuleTarball(ctx, db.Conn(ctx), InsertModuleTarballParams{
		Tarball:         tarball,
		ModuleVersionID: versionID,
	})
	return sql.Error(err)
}

func (db *pgdb) getTarball(ctx context.Context, versionID resource.ID) ([]byte, error) {
	tarball, err := q.FindModuleTarball(ctx, db.Conn(ctx), versionID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return tarball, nil
}

// toModule converts a database row into a module
func (row moduleRow) toModule() *Module {
	module := &Module{
		ID:           row.ModuleID,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		UpdatedAt:    row.UpdatedAt.Time.UTC(),
		Name:         row.Name.String,
		Provider:     row.Provider.String,
		Status:       ModuleStatus(row.Status.String),
		Organization: row.OrganizationName,
	}
	if row.RepoPath.Valid {
		module.Connection = &connections.Connection{
			VCSProviderID: row.VCSProviderID,
			Repo:          row.RepoPath.String,
		}
	}
	// versions are always maintained in descending order
	//
	// TODO: this invariant should be part of a constructor
	sort.Sort(byVersion(row.ModuleVersions))
	for i := len(row.ModuleVersions) - 1; i >= 0; i-- {
		module.Versions = append(module.Versions, ModuleVersion{
			ID:          row.ModuleVersions[i].ModuleVersionID,
			Version:     row.ModuleVersions[i].Version.String,
			CreatedAt:   row.ModuleVersions[i].CreatedAt.Time.UTC(),
			UpdatedAt:   row.ModuleVersions[i].UpdatedAt.Time.UTC(),
			ModuleID:    row.ModuleVersions[i].ModuleID,
			Status:      ModuleVersionStatus(row.ModuleVersions[i].Status.String),
			StatusError: row.ModuleVersions[i].StatusError.String,
		})
	}
	return module
}

type byVersion []ModuleVersionModel

func (v byVersion) Len() int      { return len(v) }
func (v byVersion) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v byVersion) Less(i, j int) bool {
	return semver.Compare(v[i].Version.String, v[j].Version.String) < 0
}
