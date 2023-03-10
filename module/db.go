package module

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/semver"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type (
	// pgdb is the registry database on postgres
	pgdb struct {
		otf.DB // provides access to generated SQL queries
	}

	// moduleRow is a row from a database query for modules.
	moduleRow struct {
		ModuleID         pgtype.Text            `json:"module_id"`
		CreatedAt        pgtype.Timestamptz     `json:"created_at"`
		UpdatedAt        pgtype.Timestamptz     `json:"updated_at"`
		Name             pgtype.Text            `json:"name"`
		Provider         pgtype.Text            `json:"provider"`
		Status           pgtype.Text            `json:"status"`
		OrganizationName pgtype.Text            `json:"organization_name"`
		ModuleConnection *pggen.RepoConnections `json:"module_connection"`
		Webhook          *pggen.Webhooks        `json:"webhook"`
		Versions         []pggen.ModuleVersions `json:"versions"`
	}

	// versionRow is a row from a database query for module versions.
	versionRow struct {
		ModuleVersionID pgtype.Text        `json:"module_version_id"`
		Version         pgtype.Text        `json:"version"`
		CreatedAt       pgtype.Timestamptz `json:"created_at"`
		UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
		Status          pgtype.Text        `json:"status"`
		StatusError     pgtype.Text        `json:"status_error"`
		ModuleID        pgtype.Text        `json:"module_id"`
	}
)

func (db *pgdb) CreateModule(ctx context.Context, mod *Module) error {
	params := pggen.InsertModuleParams{
		ID:               sql.String(mod.ID),
		CreatedAt:        sql.Timestamptz(mod.CreatedAt),
		UpdatedAt:        sql.Timestamptz(mod.UpdatedAt),
		Name:             sql.String(mod.Name),
		Provider:         sql.String(mod.Provider),
		Status:           sql.String(string(mod.Status)),
		OrganizationName: sql.String(mod.Organization),
	}
	_, err := db.InsertModule(ctx, params)
	return sql.Error(err)
}

func (db *pgdb) UpdateModuleStatus(ctx context.Context, moduleID string, status ModuleStatus) error {
	_, err := db.UpdateModuleStatusByID(ctx, sql.String(string(status)), sql.String(moduleID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) ListModules(ctx context.Context, opts ListModulesOptions) ([]*Module, error) {
	rows, err := db.ListModulesByOrganization(ctx, sql.String(opts.Organization))
	if err != nil {
		return nil, err
	}

	var modules []*Module
	for _, r := range rows {
		modules = append(modules, moduleRow(r).toModule())
	}
	return modules, nil
}

func (db *pgdb) GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error) {
	row, err := db.FindModuleByName(ctx, pggen.FindModuleByNameParams{
		Name:             sql.String(opts.Name),
		Provider:         sql.String(opts.Provider),
		OrganizationName: sql.String(opts.Organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) GetModuleByID(ctx context.Context, id string) (*Module, error) {
	row, err := db.FindModuleByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error) {
	row, err := db.FindModuleByWebhookID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}

	return moduleRow(row).toModule(), nil
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	_, err := db.DeleteModuleByID(ctx, sql.String(id))
	return sql.Error(err)
}

func (db *pgdb) CreateModuleVersion(ctx context.Context, version *ModuleVersion) error {
	_, err := db.InsertModuleVersion(ctx, pggen.InsertModuleVersionParams{
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

func (db *pgdb) UpdateModuleVersionStatus(ctx context.Context, opts UpdateModuleVersionStatusOptions) error {
	_, err := db.UpdateModuleVersionStatusByID(ctx, pggen.UpdateModuleVersionStatusByIDParams{
		ModuleVersionID: sql.String(opts.ID),
		Status:          sql.String(string(opts.Status)),
		StatusError:     sql.String(opts.Error),
	})
	return sql.Error(err)
}

func (db *pgdb) getModuleByVersionID(ctx context.Context, versionID string) (*Module, error) {
	row, err := db.FindModuleByModuleVersionID(ctx, sql.String(versionID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return moduleRow(row).toModule(), nil
}

func (db *pgdb) deleteModuleVersion(ctx context.Context, versionID string) error {
	_, err := db.DeleteModuleVersionByID(ctx, sql.String(versionID))
	return sql.Error(err)
}

func (db *pgdb) saveTarball(ctx context.Context, versionID string, tarball []byte) error {
	_, err := db.InsertModuleTarball(ctx, tarball, sql.String(versionID))
	return sql.Error(err)
}

func (db *pgdb) getTarball(ctx context.Context, versionID string) ([]byte, error) {
	tarball, err := db.FindModuleTarball(ctx, sql.String(versionID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return tarball, nil
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, txFunc func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return txFunc(&pgdb{tx})
	})
}

// UnmarshalModuleRow unmarshals a database row into a module
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
	if row.ModuleConnection != nil {
		module.Connection = &otf.Connection{
			VCSProviderID: row.ModuleConnection.VCSProviderID.String,
			Repo:          row.Webhook.Identifier.String,
		}
	}
	// versions are always maintained in descending order
	sort.Sort(byVersion(row.Versions))
	for i := len(row.Versions); i >= 0; i-- {
		module.Versions = append(module.Versions, ModuleVersion{
			ID:          row.Versions[i].ModuleVersionID.String,
			Version:     row.Versions[i].Version.String,
			CreatedAt:   row.Versions[i].CreatedAt.Time.UTC(),
			UpdatedAt:   row.Versions[i].UpdatedAt.Time.UTC(),
			ModuleID:    row.Versions[i].ModuleID.String,
			Status:      ModuleVersionStatus(row.Versions[i].Status.String),
			StatusError: row.Versions[i].StatusError.String,
		})
	}
	return module
}

type byVersion []pggen.ModuleVersions

func (v byVersion) Len() int      { return len(v) }
func (v byVersion) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v byVersion) Less(i, j int) bool {
	return semver.Compare(v[i].Version.String, v[j].Version.String) < 0
}
