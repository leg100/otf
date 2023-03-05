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

	// ModuleRow is a row from a database query for modules.
	ModuleRow struct {
		ModuleID         pgtype.Text            `json:"module_id"`
		CreatedAt        pgtype.Timestamptz     `json:"created_at"`
		UpdatedAt        pgtype.Timestamptz     `json:"updated_at"`
		Name             pgtype.Text            `json:"name"`
		Provider         pgtype.Text            `json:"provider"`
		Status           pgtype.Text            `json:"status"`
		OrganizationName pgtype.Text            `json:"organization_name"`
		ModuleRepo       *pggen.ModuleRepos     `json:"module_repo"`
		Webhook          *pggen.Webhooks        `json:"webhook"`
		Versions         []pggen.ModuleVersions `json:"versions"`
	}
)

func (db *pgdb) CreateModule(ctx context.Context, mod *Module) error {
	err := db.Tx(ctx, func(tx otf.DB) error {
		_, err := tx.InsertModule(ctx, pggen.InsertModuleParams{
			ID:               sql.String(mod.id),
			CreatedAt:        sql.Timestamptz(mod.createdAt),
			UpdatedAt:        sql.Timestamptz(mod.updatedAt),
			Name:             sql.String(mod.name),
			Provider:         sql.String(mod.provider),
			Status:           sql.String(string(mod.status)),
			OrganizationName: sql.String(mod.organization),
		})
		if err != nil {
			return err
		}
		if mod.connection != nil {
			_, err = tx.InsertModuleRepo(ctx, pggen.InsertModuleRepoParams{
				WebhookID:     sql.UUID(mod.connection.WebhookID),
				VCSProviderID: sql.String(mod.connection.ProviderID),
				ModuleID:      sql.String(mod.id),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (db *pgdb) UpdateModuleStatus(ctx context.Context, opts UpdateModuleStatusOptions) error {
	_, err := db.UpdateModuleStatusByID(ctx, sql.String(string(opts.Status)), sql.String(opts.ID))
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
		modules = append(modules, UnmarshalModuleRow(ModuleRow(r)))
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

	return UnmarshalModuleRow(ModuleRow(row)), nil
}

func (db *pgdb) GetModuleByID(ctx context.Context, id string) (*Module, error) {
	row, err := db.FindModuleByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}

	return UnmarshalModuleRow(ModuleRow(row)), nil
}

func (db *pgdb) GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error) {
	row, err := db.FindModuleByWebhookID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}

	return UnmarshalModuleRow(ModuleRow(row)), nil
}

func (db *pgdb) DeleteModule(ctx context.Context, id string) error {
	return deleteModule(ctx, db, id)
}

func (db *pgdb) CreateModuleVersion(ctx context.Context, version *ModuleVersion) error {
	_, err := db.InsertModuleVersion(ctx, pggen.InsertModuleVersionParams{
		ModuleVersionID: sql.String(version.id),
		Version:         sql.String(version.version),
		CreatedAt:       sql.Timestamptz(version.createdAt),
		UpdatedAt:       sql.Timestamptz(version.updatedAt),
		ModuleID:        sql.String(version.moduleID),
		Status:          sql.String(string(version.status)),
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *pgdb) UpdateModuleVersionStatus(ctx context.Context, opts UpdateModuleVersionStatusOptions) (*ModuleVersion, error) {
	row, err := db.UpdateModuleVersionStatusByID(ctx, pggen.UpdateModuleVersionStatusByIDParams{
		ModuleVersionID: sql.String(opts.ID),
		Status:          sql.String(string(opts.Status)),
		StatusError:     sql.String(opts.Error),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalModuleVersionRow(ModuleVersionRow(row)), nil
}

func (db *pgdb) UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) error {
	_, err := db.InsertModuleTarball(ctx, opts.Tarball, sql.String(opts.VersionID))
	return sql.Error(err)
}

func (db *pgdb) DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error) {
	tarball, err := db.FindModuleTarball(ctx, sql.String(opts.ModuleVersionID))
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

func deleteModule(ctx context.Context, db otf.Database, id string) error {
	_, err := db.DeleteModuleByID(ctx, sql.String(id))
	return sql.Error(err)
}

// UnmarshalModuleRow unmarshals a database row into a module
func UnmarshalModuleRow(row ModuleRow) *Module {
	module := &Module{
		id:           row.ModuleID.String,
		createdAt:    row.CreatedAt.Time.UTC(),
		updatedAt:    row.UpdatedAt.Time.UTC(),
		name:         row.Name.String,
		provider:     row.Provider.String,
		status:       ModuleStatus(row.Status.String),
		organization: row.OrganizationName.String,
	}
	if row.ModuleRepo != nil {
		module.connection = &connection{
			ProviderID: row.ModuleRepo.VCSProviderID.String,
			WebhookID:  row.Webhook.WebhookID.Bytes,
			Identifier: row.Webhook.Identifier.String,
		}
	}
	for _, version := range row.Versions {
		module.versions[version.Version.String] = UnmarshalModuleVersionRow(ModuleVersionRow(version))
	}
	return module
}
