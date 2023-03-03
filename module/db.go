package module

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is the registry database on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func (db *pgdb) CreateModule(ctx context.Context, mod *Module) error {
	return createModule(ctx, db, mod)
}

func createModule(ctx context.Context, db otf.Database, mod *Module) error {
	err := db.Transaction(ctx, func(tx otf.Database) error {
		_, err := tx.InsertModule(ctx, pggen.InsertModuleParams{
			ID:               sql.String(mod.ID),
			CreatedAt:        sql.Timestamptz(mod.CreatedAt()),
			UpdatedAt:        sql.Timestamptz(mod.UpdatedAt()),
			Name:             sql.String(mod.Name()),
			Provider:         sql.String(mod.Provider()),
			Status:           sql.String(string(mod.Status())),
			OrganizationName: sql.String(mod.Organization()),
		})
		if err != nil {
			return err
		}
		if mod.Repo() != nil {
			_, err = tx.InsertModuleRepo(ctx, pggen.InsertModuleRepoParams{
				WebhookID:     sql.UUID(mod.Repo().WebhookID),
				VCSProviderID: sql.String(mod.Repo().ProviderID),
				ModuleID:      sql.String(mod.ID),
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
		ModuleVersionID: sql.String(version.ID),
		Version:         sql.String(version.Version()),
		CreatedAt:       sql.Timestamptz(version.CreatedAt()),
		UpdatedAt:       sql.Timestamptz(version.UpdatedAt()),
		ModuleID:        sql.String(version.ModuleID),
		Status:          sql.String(string(version.Status())),
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
	return uploadModuleVersion(ctx, db, opts)
}

func (db *pgdb) DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error) {
	tarball, err := db.FindModuleTarball(ctx, sql.String(opts.ModuleVersionID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return tarball, nil
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(db) error) error {
	return db.Transaction(ctx, func(tx otf.Database) error {
		return callback(newPGDB(tx))
	})
}

func uploadModuleVersion(ctx context.Context, tx otf.Database, opts UploadModuleVersionOptions) error {
	_, err := tx.InsertModuleTarball(ctx, opts.Tarball, sql.String(opts.ModuleVersionID))
	return sql.Error(err)
}

func deleteModule(ctx context.Context, db otf.Database, id string) error {
	_, err := db.DeleteModuleByID(ctx, sql.String(id))
	return sql.Error(err)
}
