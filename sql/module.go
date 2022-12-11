package sql

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateModule(ctx context.Context, mod *otf.Module) error {
	_, err := db.InsertModule(ctx, pggen.InsertModuleParams{
		ID:             String(mod.ID()),
		CreatedAt:      Timestamptz(mod.CreatedAt()),
		UpdatedAt:      Timestamptz(mod.UpdatedAt()),
		OrganizationID: String(mod.Organization().ID()),
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateModuleVersion(ctx context.Context, version *otf.ModuleVersion) error {
	_, err := db.InsertModuleVersion(ctx, pggen.InsertModuleVersionParams{
		ModuleVersionID: String(version.ID()),
		Version:         String(version.Version()),
		CreatedAt:       Timestamptz(version.CreatedAt()),
		UpdatedAt:       Timestamptz(version.UpdatedAt()),
		ModuleID:        String(version.ModuleID()),
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UploadModuleVersion(ctx context.Context, opts otf.UploadModuleVersionOptions) error {
	_, err := db.InsertModuleTarball(ctx, opts.Tarball, String(opts.ModuleVersionID))
	if err != nil {
		return err
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
		OrganizatonName: String(opts.Organization.Name()),
	})
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalModuleRow(otf.ModuleRow(row)), nil
}

func (db *DB) GetModuleByWebhook(ctx context.Context, id uuid.UUID) (*otf.Module, error) {
	row, err := db.FindModuleByWebhookID(ctx, UUID(id))
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalModuleRow(otf.ModuleRow(row)), nil
}

func (db *DB) DownloadModuleVersion(ctx context.Context, opts otf.DownloadModuleOptions) ([]byte, error) {
	tarball, err := db.FindModuleTarball(ctx, String(opts.ModuleVersionID))
	if err != nil {
		return nil, databaseError(err)
	}
	return tarball, nil
}
