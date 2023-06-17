package state

import (
	"context"
	"errors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	db interface {
		internal.DB

		createVersion(context.Context, *Version) error
		listVersions(ctx context.Context, workspaceID string, opts internal.ListOptions) (*VersionList, error)
		getVersion(ctx context.Context, svID string) (*Version, error)
		getCurrentVersion(ctx context.Context, workspaceID string) (*Version, error)
		getState(ctx context.Context, versionID string) ([]byte, error)
		getOutput(ctx context.Context, outputID string) (*Output, error)
		updateCurrentVersion(context.Context, string, string) error
		deleteVersion(ctx context.Context, versionID string) error
		tx(context.Context, func(db) error) error
	}

	// pgdb is a state/state-version database on postgres
	pgdb struct {
		internal.DB // provides access to generated SQL queries
	}

	// pgRow is a row from a postgres query for a state version.
	pgRow struct {
		StateVersionID      pgtype.Text                 `json:"state_version_id"`
		CreatedAt           pgtype.Timestamptz          `json:"created_at"`
		Serial              int                         `json:"serial"`
		State               []byte                      `json:"state"`
		WorkspaceID         pgtype.Text                 `json:"workspace_id"`
		StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
	}
)

func (db *pgdb) createVersion(ctx context.Context, v *Version) error {
	return db.Tx(ctx, func(db internal.DB) error {
		_, err := db.InsertStateVersion(ctx, pggen.InsertStateVersionParams{
			ID:          sql.String(v.ID),
			CreatedAt:   sql.Timestamptz(v.CreatedAt),
			Serial:      int(v.Serial),
			State:       v.State,
			WorkspaceID: sql.String(v.WorkspaceID),
		})
		if err != nil {
			return err
		}

		for _, svo := range v.Outputs {
			_, err := db.InsertStateVersionOutput(ctx, pggen.InsertStateVersionOutputParams{
				ID:             sql.String(svo.ID),
				Name:           sql.String(svo.Name),
				Sensitive:      svo.Sensitive,
				Type:           sql.String(svo.Type),
				Value:          sql.String(svo.Value),
				StateVersionID: sql.String(v.ID),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) listVersions(ctx context.Context, workspaceID string, opts internal.ListOptions) (*VersionList, error) {
	batch := &pgx.Batch{}

	db.FindStateVersionsByWorkspaceIDBatch(batch, pggen.FindStateVersionsByWorkspaceIDParams{
		WorkspaceID: sql.String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	db.CountStateVersionsByWorkspaceIDBatch(batch, sql.String(workspaceID))

	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindStateVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountStateVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}

	var items []*Version
	for _, r := range rows {
		items = append(items, pgRow(r).toVersion())
	}

	return &VersionList{
		Items:      items,
		Pagination: internal.NewPagination(opts, count),
	}, nil
}

func (db *pgdb) getVersion(ctx context.Context, svID string) (*Version, error) {
	result, err := db.FindStateVersionByID(ctx, sql.String(svID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toVersion(), nil
}

func (db *pgdb) getCurrentVersion(ctx context.Context, workspaceID string) (*Version, error) {
	result, err := db.FindCurrentStateVersionByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toVersion(), nil
}

func (db *pgdb) getState(ctx context.Context, id string) ([]byte, error) {
	return db.FindStateVersionStateByID(ctx, sql.String(id))
}

// deleteVersion deletes a state version from the DB
func (db *pgdb) deleteVersion(ctx context.Context, id string) error {
	_, err := db.DeleteStateVersionByID(ctx, sql.String(id))
	if err != nil {
		err = sql.Error(err)
		var fkerr *internal.ForeignKeyError
		if errors.As(err, &fkerr) {
			if fkerr.ConstraintName == "current_state_version_id_fk" && fkerr.TableName == "workspaces" {
				return ErrCurrentVersionDeletionAttempt
			}
		}
		return err
	}
	return nil
}

func (db *pgdb) updateCurrentVersion(ctx context.Context, workspaceID, svID string) error {
	_, err := db.UpdateWorkspaceCurrentStateVersionID(ctx, sql.String(svID), sql.String(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) tx(ctx context.Context, txfunc func(db) error) error {
	return db.Tx(ctx, func(tx internal.DB) error {
		return txfunc(&pgdb{tx})
	})
}

func (row pgRow) toVersion() *Version {
	sv := Version{
		ID:          row.StateVersionID.String,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Serial:      int64(row.Serial),
		State:       row.State,
		WorkspaceID: row.WorkspaceID.String,
		Outputs:     make(OutputList, len(row.StateVersionOutputs)),
	}
	for _, r := range row.StateVersionOutputs {
		sv.Outputs[r.Name.String] = outputRow(r).toOutput()
	}
	return &sv
}
