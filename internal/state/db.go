package state

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type (
	// pgdb is a state/state-version database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// pgRow is a row from a postgres query for a state version.
	pgRow struct {
		StateVersionID      pgtype.Text               `json:"state_version_id"`
		CreatedAt           pgtype.Timestamptz        `json:"created_at"`
		Serial              pgtype.Int4               `json:"serial"`
		State               []byte                    `json:"state"`
		WorkspaceID         pgtype.Text               `json:"workspace_id"`
		Status              pgtype.Text               `json:"status"`
		StateVersionOutputs []sqlc.StateVersionOutput `json:"state_version_outputs"`
	}
)

func (row pgRow) toVersion() *Version {
	sv := Version{
		ID:          row.StateVersionID.String,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Serial:      int64(row.Serial.Int32),
		State:       row.State,
		Status:      Status(row.Status.String),
		WorkspaceID: row.WorkspaceID.String,
		Outputs:     make(map[string]*Output, len(row.StateVersionOutputs)),
	}
	for _, r := range row.StateVersionOutputs {
		sv.Outputs[r.Name.String] = outputRow(r).toOutput()
	}
	return &sv
}

func (db *pgdb) createVersion(ctx context.Context, v *Version) error {
	return db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		err := q.InsertStateVersion(ctx, sqlc.InsertStateVersionParams{
			ID:          sql.ID(v.ID),
			CreatedAt:   sql.Timestamptz(v.CreatedAt),
			Serial:      sql.Int4(int(v.Serial)),
			State:       v.State,
			Status:      sql.String(string(v.Status)),
			WorkspaceID: sql.ID(v.WorkspaceID),
		})
		if err != nil {
			return err
		}

		for _, svo := range v.Outputs {
			err := q.InsertStateVersionOutput(ctx, sqlc.InsertStateVersionOutputParams{
				ID:             sql.ID(svo.ID),
				Name:           sql.String(svo.Name),
				Sensitive:      sql.Bool(svo.Sensitive),
				Type:           sql.String(svo.Type),
				Value:          svo.Value,
				StateVersionID: sql.ID(v.ID),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) createOutputs(ctx context.Context, outputs []*Output) error {
	return db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		for _, svo := range outputs {
			err := q.InsertStateVersionOutput(ctx, sqlc.InsertStateVersionOutputParams{
				ID:             sql.ID(svo.ID),
				Name:           sql.String(svo.Name),
				Sensitive:      sql.Bool(svo.Sensitive),
				Type:           sql.String(svo.Type),
				Value:          svo.Value,
				StateVersionID: sql.ID(svo.StateVersionID),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) uploadStateAndFinalize(ctx context.Context, svID resource.ID, state []byte) error {
	err := db.Querier(ctx).UpdateState(ctx, sqlc.UpdateStateParams{
		State:          state,
		StateVersionID: sql.ID(svID),
	})
	return sql.Error(err)
}

func (db *pgdb) listVersions(ctx context.Context, workspaceID resource.ID, opts resource.PageOptions) (*resource.Page[*Version], error) {
	q := db.Querier(ctx)

	rows, err := q.FindStateVersionsByWorkspaceID(ctx, sqlc.FindStateVersionsByWorkspaceIDParams{
		WorkspaceID: sql.ID(workspaceID),
		Limit:       sql.GetLimit(opts.PageOptions),
		Offset:      sql.GetOffset(opts),
	})
	if err != nil {
		return nil, err
	}

	count, err := q.CountStateVersionsByWorkspaceID(ctx, sql.ID(workspaceID))
	if err != nil {
		return nil, err
	}

	items := make([]*Version, len(rows))
	for i, r := range rows {
		items[i] = pgRow(r).toVersion()
	}
	return resource.NewPage(items, opts, internal.Int64(count)), nil
}

func (db *pgdb) getVersion(ctx context.Context, svID resource.ID) (*Version, error) {
	result, err := db.Querier(ctx).FindStateVersionByID(ctx, sql.ID(svID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toVersion(), nil
}

func (db *pgdb) getVersionForUpdate(ctx context.Context, svID resource.ID) (*Version, error) {
	result, err := db.Querier(ctx).FindStateVersionByIDForUpdate(ctx, sql.ID(svID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toVersion(), nil
}

func (db *pgdb) getCurrentVersion(ctx context.Context, workspaceID resource.ID) (*Version, error) {
	result, err := db.Querier(ctx).FindCurrentStateVersionByWorkspaceID(ctx, sql.ID(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toVersion(), nil
}

func (db *pgdb) getState(ctx context.Context, id resource.ID) ([]byte, error) {
	return db.Querier(ctx).FindStateVersionStateByID(ctx, sql.ID(id))
}

// deleteVersion deletes a state version from the DB
func (db *pgdb) deleteVersion(ctx context.Context, id resource.ID) error {
	_, err := db.Querier(ctx).DeleteStateVersionByID(ctx, sql.ID(id))
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

func (db *pgdb) updateCurrentVersion(ctx context.Context, workspaceID, svID resource.ID) error {
	_, err := db.Querier(ctx).UpdateWorkspaceCurrentStateVersionID(ctx, sqlc.UpdateWorkspaceCurrentStateVersionIDParams{
		StateVersionID: sql.ID(svID),
		WorkspaceID:    sql.ID(workspaceID),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) discardPending(ctx context.Context, workspaceID resource.ID) error {
	err := db.Querier(ctx).DiscardPendingStateVersionsByWorkspaceID(ctx, sql.ID(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
