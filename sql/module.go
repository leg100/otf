package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
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

func (db *DB) UploadModuleVersion(ctx context.Context, opts otf.UploadModuleVersionOptions) error {
	result, err := db.Querier.UploadModuleVersion(ctx, pggen.UploadModuleVersionParams{
		Tarball   : opts.Tarball,
		UpdatedAt :
		ModuleID  :
		Version   :
	})
	cv, err := otf.UnmarshalModuleResult(otf.ModuleResult(result))
	if err != nil {
		return err
	}

	if err := fn(cv, newConfigUploader(tx, cv.ID())); err != nil {
		return err
	}
	return nil
}

func (db *DB) ListModules(ctx context.Context, workspaceID string, opts otf.ListModulesOptions) (*otf.ModuleList, error) {
	batch := &pgx.Batch{}
	db.FindModulesByWorkspaceIDBatch(batch, pggen.FindModulesByWorkspaceIDParams{
		WorkspaceID: String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	db.CountModulesByWorkspaceIDBatch(batch, String(workspaceID))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindModulesByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountModulesByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}

	var items []*otf.Module
	for _, r := range rows {
		cv, err := otf.UnmarshalModuleResult(otf.ModuleResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, cv)
	}

	return &otf.ModuleList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *DB) GetModule(ctx context.Context, opts otf.ModuleGetOptions) (*otf.Module, error) {
	if opts.ID != nil {
		result, err := db.FindModuleByID(ctx, String(*opts.ID))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalModuleResult(otf.ModuleResult(result))
	} else if opts.WorkspaceID != nil {
		result, err := db.FindModuleLatestByWorkspaceID(ctx, String(*opts.WorkspaceID))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalModuleResult(otf.ModuleResult(result))
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}

func (db *DB) GetConfig(ctx context.Context, id string) ([]byte, error) {
	return db.DownloadModule(ctx, String(id))
}

func (db *DB) DeleteModule(ctx context.Context, id string) error {
	_, err := db.DeleteModuleByID(ctx, String(id))
	if err != nil {
		return databaseError(err)
	}
	return nil
}

func (db *DB) insertCVStatusTimestamp(ctx context.Context, cv *otf.Module) error {
	sts, err := cv.StatusTimestamp(cv.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertModuleStatusTimestamp(ctx, pggen.InsertModuleStatusTimestampParams{
		ID:        String(cv.ID()),
		Status:    String(string(cv.Status())),
		Timestamp: Timestamptz(sts),
	})
	return err
}
