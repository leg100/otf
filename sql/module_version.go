package sql

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateModuleVersion(ctx context.Context, version *otf.ModuleVersion) error {
	_, err := db.InsertModuleVersion(ctx, pggen.InsertModuleVersionParams{
		ModuleVersionID: String(version.ID()),
		Version:         String(version.Version()),
		CreatedAt:       Timestamptz(version.CreatedAt()),
		UpdatedAt:       Timestamptz(version.UpdatedAt()),
		ModuleID:        String(version.ModuleID()),
		Status:          String(string(version.Status())),
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateModuleVersionStatus(ctx context.Context, opts otf.UpdateModuleVersionStatusOptions) (*otf.ModuleVersion, error) {
	row, err := db.Querier.UpdateModuleVersionStatus(ctx, pggen.UpdateModuleVersionStatusParams{
		ModuleVersionID: String(opts.ID),
		Status:          String(string(opts.Status)),
		StatusError:     String(opts.Error),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalModuleVersionRow(otf.ModuleVersionRow(row)), nil
}

func (db *DB) UploadModuleVersion(ctx context.Context, opts otf.UploadModuleVersionOptions) error {
	_, err := db.InsertModuleTarball(ctx, opts.Tarball, String(opts.ModuleVersionID))
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) DownloadModuleVersion(ctx context.Context, opts otf.DownloadModuleOptions) ([]byte, error) {
	tarball, err := db.FindModuleTarball(ctx, String(opts.ModuleVersionID))
	if err != nil {
		return nil, databaseError(err)
	}
	return tarball, nil
}
