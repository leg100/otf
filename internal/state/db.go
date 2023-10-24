package state

import (
	"context"
	"errors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a state/state-version database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// pgRow is a row from a postgres query for a state version.
	pgRow struct {
		StateVersionID      pgtype.Text                 `json:"state_version_id"`
		CreatedAt           pgtype.Timestamptz          `json:"created_at"`
		Serial              pgtype.Int4                 `json:"serial"`
		State               []byte                      `json:"state"`
		WorkspaceID         pgtype.Text                 `json:"workspace_id"`
		Status              pgtype.Text                 `json:"status"`
		StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
	}
)

func (row pgRow) toVersion() *Version {
	sv := Version{
		ID:          row.StateVersionID.String,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Serial:      int64(row.Serial.Int),
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
	return db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertStateVersion(ctx, pggen.InsertStateVersionParams{
			ID:          sql.String(v.ID),
			CreatedAt:   sql.Timestamptz(v.CreatedAt),
			Serial:      sql.Int4(int(v.Serial)),
			State:       v.State,
			Status:      sql.String(string(v.Status)),
			WorkspaceID: sql.String(v.WorkspaceID),
		})
		if err != nil {
			return err
		}

		for _, svo := range v.Outputs {
			_, err := q.InsertStateVersionOutput(ctx, pggen.InsertStateVersionOutputParams{
				ID:             sql.String(svo.ID),
				Name:           sql.String(svo.Name),
				Sensitive:      svo.Sensitive,
				Type:           sql.String(svo.Type),
				Value:          svo.Value,
				StateVersionID: sql.String(v.ID),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// TODO: rename this updateStateAndFinalize
func (db *pgdb) updateState(ctx context.Context, svID string, state []byte) error {
	_, err := db.Conn(ctx).UpdateState(ctx, state, sql.String(svID))
	return sql.Error(err)
}

func (db *pgdb) listVersions(ctx context.Context, workspaceID string, opts resource.PageOptions) (*resource.Page[*Version], error) {
	q := db.Conn(ctx)
	batch := &pgx.Batch{}

	q.FindStateVersionsByWorkspaceIDBatch(batch, pggen.FindStateVersionsByWorkspaceIDParams{
		WorkspaceID: sql.String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	q.CountStateVersionsByWorkspaceIDBatch(batch, sql.String(workspaceID))

	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindStateVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountStateVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}

	var items []*Version
	for _, r := range rows {
		items = append(items, pgRow(r).toVersion())
	}

	return resource.NewPage(items, opts, internal.Int64(count.Int)), nil
}

func (db *pgdb) getVersion(ctx context.Context, svID string) (*Version, error) {
	result, err := db.Conn(ctx).FindStateVersionByID(ctx, sql.String(svID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toVersion(), nil
}

func (db *pgdb) getCurrentVersion(ctx context.Context, workspaceID string) (*Version, error) {
	result, err := db.Conn(ctx).FindCurrentStateVersionByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toVersion(), nil
}

func (db *pgdb) getState(ctx context.Context, id string) ([]byte, error) {
	return db.Conn(ctx).FindStateVersionStateByID(ctx, sql.String(id))
}

// deleteVersion deletes a state version from the DB
func (db *pgdb) deleteVersion(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteStateVersionByID(ctx, sql.String(id))
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
	_, err := db.Conn(ctx).UpdateWorkspaceCurrentStateVersionID(ctx, sql.String(svID), sql.String(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
