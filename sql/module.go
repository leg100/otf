package sql

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateModule(ctx context.Context, mod *otf.Module) error {
	return CreateModule(ctx, db, mod)
}

func CreateModule(ctx context.Context, db otf.Database, mod *otf.Module) error {
	err := db.Transaction(ctx, func(tx otf.Database) error {
		_, err := tx.InsertModule(ctx, pggen.InsertModuleParams{
			ID:             String(mod.ID()),
			CreatedAt:      Timestamptz(mod.CreatedAt()),
			UpdatedAt:      Timestamptz(mod.UpdatedAt()),
			Name:           String(mod.Name()),
			Provider:       String(mod.Provider()),
			Status:         String(string(mod.Status())),
			OrganizationID: String(mod.Organization().ID()),
		})
		if err != nil {
			return err
		}
		if mod.Repo() != nil {
			_, err = tx.InsertModuleRepo(ctx, pggen.InsertModuleRepoParams{
				WebhookID:     UUID(mod.Repo().WebhookID),
				VCSProviderID: String(mod.Repo().ProviderID),
				ModuleID:      String(mod.ID()),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return Error(err)
}

func (db *DB) UpdateModuleStatus(ctx context.Context, opts otf.UpdateModuleStatusOptions) error {
	_, err := db.Querier.UpdateModuleStatusByID(ctx, String(string(opts.Status)), String(opts.ID))
	if err != nil {
		return Error(err)
	}
	return nil
}

func (db *DB) ListModules(ctx context.Context, opts otf.ListModulesOptions) ([]*otf.Module, error) {
	rows, err := db.ListModulesByOrganization(ctx, String(opts.Organization))
	if err != nil {
		return nil, err
	}

	var modules []*otf.Module
	for _, r := range rows {
		modules = append(modules, otf.UnmarshalModuleRow(otf.ModuleRow(r)))
	}
	return modules, nil
}

func (db *DB) GetModule(ctx context.Context, opts otf.GetModuleOptions) (*otf.Module, error) {
	row, err := db.FindModuleByName(ctx, pggen.FindModuleByNameParams{
		Name:            String(opts.Name),
		Provider:        String(opts.Provider),
		OrganizatonName: String(opts.Organization),
	})
	if err != nil {
		return nil, Error(err)
	}

	return otf.UnmarshalModuleRow(otf.ModuleRow(row)), nil
}

func (db *DB) GetModuleByID(ctx context.Context, id string) (*otf.Module, error) {
	row, err := db.FindModuleByID(ctx, String(id))
	if err != nil {
		return nil, Error(err)
	}

	return otf.UnmarshalModuleRow(otf.ModuleRow(row)), nil
}

func (db *DB) GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*otf.Module, error) {
	row, err := db.FindModuleByWebhookID(ctx, UUID(id))
	if err != nil {
		return nil, Error(err)
	}

	return otf.UnmarshalModuleRow(otf.ModuleRow(row)), nil
}

func (db *DB) DeleteModule(ctx context.Context, id string) error {
	_, err := db.DeleteModuleByID(ctx, String(id))
	if err != nil {
		return Error(err)
	}
	return nil
}
