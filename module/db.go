package module

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
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
		Latest           pgtype.Text            `json:"latest"`
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

func (db *pgdb) CreateModule(ctx context.Context, mod *otf.Module) error {
	params := pggen.InsertModuleParams{
		ID:               sql.String(mod.ID),
		CreatedAt:        sql.Timestamptz(mod.CreatedAt),
		UpdatedAt:        sql.Timestamptz(mod.UpdatedAt),
		Name:             sql.String(mod.Name),
		Provider:         sql.String(mod.Provider),
		Status:           sql.String(string(mod.Status)),
		OrganizationName: sql.String(mod.Organization),
	}
	if mod.Latest != nil {
		params.Latest = sql.String(mod.Latest.ID)
	}
	_, err := db.InsertModule(ctx, params)
	return sql.Error(err)
}

func (db *pgdb) UpdateModuleStatus(ctx context.Context, moduleID string, status otf.ModuleStatus) error {
	_, err := db.UpdateModuleStatusByID(ctx, sql.String(string(status)), sql.String(moduleID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) ListModules(ctx context.Context, opts ListModulesOptions) ([]*otf.Module, error) {
	rows, err := db.ListModulesByOrganization(ctx, sql.String(opts.Organization))
	if err != nil {
		return nil, err
	}

	var modules []*otf.Module
	for _, r := range rows {
		modules = append(modules, UnmarshalModuleRow(moduleRow(r)))
	}
	return modules, nil
}

func (db *pgdb) GetModule(ctx context.Context, opts GetModuleOptions) (*otf.Module, error) {
	row, err := db.FindModuleByName(ctx, pggen.FindModuleByNameParams{
		Name:             sql.String(opts.Name),
		Provider:         sql.String(opts.Provider),
		OrganizationName: sql.String(opts.Organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}

	return UnmarshalModuleRow(moduleRow(row)), nil
}

func (db *pgdb) GetModuleByID(ctx context.Context, id string) (*otf.Module, error) {
	row, err := db.FindModuleByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}

	return UnmarshalModuleRow(moduleRow(row)), nil
}

func (db *pgdb) GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*otf.Module, error) {
	row, err := db.FindModuleByWebhookID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}

	return UnmarshalModuleRow(moduleRow(row)), nil
}

func (db *pgdb) DeleteModule(ctx context.Context, id string) error {
	_, err := db.DeleteModuleByID(ctx, sql.String(id))
	return sql.Error(err)
}

func (db *pgdb) CreateModuleVersion(ctx context.Context, version *otf.ModuleVersion) error {
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

func (db *pgdb) UpdateModuleVersionStatus(ctx context.Context, opts otf.UpdateModuleVersionStatusOptions) (*otf.ModuleVersion, error) {
	row, err := db.UpdateModuleVersionStatusByID(ctx, pggen.UpdateModuleVersionStatusByIDParams{
		ModuleVersionID: sql.String(opts.ID),
		Status:          sql.String(string(opts.Status)),
		StatusError:     sql.String(opts.Error),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalModuleVersionRow(versionRow(row)), nil
}

func (db *pgdb) getModuleByVersionID(ctx context.Context, versionID string) (*otf.Module, error) {
	row, err := db.FindModuleByModuleVersionID(ctx, sql.String(versionID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalModuleRow(moduleRow(row)), nil
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

func (db *pgdb) updateLatest(ctx context.Context, moduleID string, modver *otf.ModuleVersion) error {
	latestID := pgtype.Text{Status: pgtype.Null}
	if modver != nil {
		latestID = sql.String(modver.ID)
	}
	_, err := db.UpdateModuleLatestVersionByID(ctx, latestID, sql.String(moduleID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, txFunc func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return txFunc(&pgdb{tx})
	})
}

// UnmarshalModuleRow unmarshals a database row into a module
func UnmarshalModuleRow(row moduleRow) *otf.Module {
	module := &otf.Module{
		ID:           row.ModuleID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		UpdatedAt:    row.UpdatedAt.Time.UTC(),
		Name:         row.Name.String,
		Provider:     row.Provider.String,
		Status:       otf.ModuleStatus(row.Status.String),
		Organization: row.OrganizationName.String,
	}
	if row.ModuleConnection != nil {
		module.Connection = &otf.Connection{
			VCSProviderID: row.ModuleConnection.VCSProviderID.String,
			RepoID:        row.Webhook.WebhookID.Bytes,
			Identifier:    row.Webhook.Identifier.String,
		}
	}
	for _, version := range row.Versions {
		module.Versions[version.Version.String] = UnmarshalModuleVersionRow(versionRow(version))
	}
	return module
}

// UnmarshalModuleVersionRow unmarshals a database row into a module version
func UnmarshalModuleVersionRow(row versionRow) *otf.ModuleVersion {
	return &otf.ModuleVersion{
		ID:          row.ModuleVersionID.String,
		Version:     row.Version.String,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		UpdatedAt:   row.UpdatedAt.Time.UTC(),
		ModuleID:    row.ModuleID.String,
		Status:      otf.ModuleVersionStatus(row.Status.String),
		StatusError: row.StatusError.String,
	}
}
