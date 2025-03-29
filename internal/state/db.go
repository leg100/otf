package state

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type (
	// pgdb is a state/state-version database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// versionModel is the database model for a state version row.
	versionModel struct {
		StateVersionID      resource.TfeID `db:"state_version_id"`
		CreatedAt           time.Time      `db:"created_at"`
		Serial              int64          `db:"serial"`
		State               []byte         `db:"state"`
		WorkspaceID         resource.TfeID `db:"workspace_id"`
		Status              Status         `db:"status"`
		StateVersionOutputs []outputModel  `db:"state_version_outputs"`
	}
)

func (db *pgdb) createVersion(ctx context.Context, v *Version) error {
	return db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		_, err := db.Exec(ctx, `
INSERT INTO state_versions (
    state_version_id,
    created_at,
    serial,
    state,
    status,
    workspace_id
) VALUES (
    @id,
    @created_at,
    @serial,
    @state,
	@status,
	@workspace_id
)`, pgx.NamedArgs{
			"id":           v.ID,
			"created_at":   v.CreatedAt,
			"serial":       v.Serial,
			"state":        v.State,
			"status":       v.Status,
			"workspace_id": v.WorkspaceID,
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *pgdb) createOutputs(ctx context.Context, outputs []*Output) error {
	return db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		for _, svo := range outputs {
			_, err := db.Exec(ctx, `
INSERT INTO state_version_outputs (
    state_version_output_id,
    name,
    sensitive,
    type,
    value,
    state_version_id
) VALUES (
    @state_version_output_id,
    @name,
    @sensitive,
    @type,
	@value,
	@state_version_id
)`, pgx.NamedArgs{
				"state_version_output_id": svo.ID,
				"name":                    svo.Name,
				"sensitive":               svo.Sensitive,
				"type":                    svo.Type,
				"value":                   svo.Value,
				"state_version_id":        svo.StateVersionID,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) uploadStateAndFinalize(ctx context.Context, svID resource.ID, state []byte) error {
	_, err := db.Exec(ctx, `
UPDATE state_versions
SET state = $1, status = 'finalized'
WHERE state_version_id = $2
`, state, svID)
	return err
}

func (db *pgdb) listVersions(ctx context.Context, workspaceID resource.ID, opts resource.PageOptions) (*resource.Page[*Version], error) {
	rows := db.Query(ctx, `
SELECT
    sv.state_version_id, sv.created_at, sv.serial, sv.state, sv.workspace_id, sv.status,
    (
        SELECT array_agg(svo.*)::state_version_outputs[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
        GROUP BY svo.state_version_id
    ) AS state_version_outputs
FROM state_versions sv
WHERE sv.workspace_id = $1
AND   sv.status = 'finalized'
ORDER BY created_at DESC
LIMIT $2::int
OFFSET $3::int
`, workspaceID, sql.GetLimit(opts), sql.GetOffset(opts))
	items, err := sql.CollectRows(rows, db.scanVersion)
	if err != nil {
		return nil, err
	}

	count, err := db.Int(ctx, `
SELECT count(*)
FROM state_versions
WHERE workspace_id = $1
AND status = 'finalized'
`, workspaceID)
	if err != nil {
		return nil, err
	}

	return resource.NewPage(items, opts, &count), nil
}

func (db *pgdb) scanVersion(row pgx.CollectableRow) (*Version, error) {
	model, err := pgx.RowToStructByName[versionModel](row)
	if err != nil {
		return nil, err
	}
	sv := Version{
		ID:          model.StateVersionID,
		CreatedAt:   model.CreatedAt,
		Serial:      model.Serial,
		State:       model.State,
		Status:      model.Status,
		WorkspaceID: model.WorkspaceID,
		Outputs:     make(map[string]*Output, len(model.StateVersionOutputs)),
	}
	for _, output := range model.StateVersionOutputs {
		sv.Outputs[output.Name] = output.toOutput()
	}
	return &sv, nil

}

func (db *pgdb) getVersion(ctx context.Context, svID resource.ID) (*Version, error) {
	rows := db.Query(ctx, `
SELECT
    sv.state_version_id, sv.created_at, sv.serial, sv.state, sv.workspace_id, sv.status,
    (
        SELECT array_agg(svo.*)::state_version_outputs[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
        GROUP BY svo.state_version_id
    ) AS state_version_outputs
FROM state_versions sv
WHERE sv.state_version_id = $1
`, svID)
	return sql.CollectOneRow(rows, db.scanVersion)
}

func (db *pgdb) getVersionForUpdate(ctx context.Context, svID resource.ID) (*Version, error) {
	rows := db.Query(ctx, `
SELECT
    sv.state_version_id, sv.created_at, sv.serial, sv.state, sv.workspace_id, sv.status,
    (
        SELECT array_agg(svo.*)::state_version_outputs[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
        GROUP BY svo.state_version_id
    ) AS state_version_outputs
FROM state_versions sv
WHERE sv.state_version_id = $1
FOR UPDATE OF sv
`, svID)
	return sql.CollectOneRow(rows, db.scanVersion)
}

func (db *pgdb) getCurrentVersion(ctx context.Context, workspaceID resource.ID) (*Version, error) {
	rows := db.Query(ctx, `
SELECT
    sv.state_version_id, sv.created_at, sv.serial, sv.state, sv.workspace_id, sv.status,
    (
        SELECT array_agg(svo.*)::state_version_outputs[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
        GROUP BY svo.state_version_id
    ) AS state_version_outputs
FROM state_versions sv
JOIN workspaces w ON w.current_state_version_id = sv.state_version_id
WHERE w.workspace_id = $1
`, workspaceID)
	return sql.CollectOneRow(rows, db.scanVersion)
}

func (db *pgdb) getState(ctx context.Context, id resource.ID) ([]byte, error) {
	rows := db.Query(ctx, `
SELECT state
FROM state_versions
WHERE state_version_id = $1
`, id)
	return sql.CollectOneRow(rows, pgx.RowTo[[]byte])
}

// deleteVersion deletes a state version from the DB
func (db *pgdb) deleteVersion(ctx context.Context, id resource.ID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM state_versions
WHERE state_version_id = $1
`, id)
	if err != nil {
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
	_, err := db.Exec(ctx, `
UPDATE workspaces
SET current_state_version_id = $1
WHERE workspace_id = $2
RETURNING workspace_id
`, svID, workspaceID)
	return err
}

func (db *pgdb) discardAnyPending(ctx context.Context, workspaceID resource.ID) error {
	_, err := db.Exec(ctx, `
UPDATE state_versions
SET status = 'discarded'
WHERE workspace_id = $1
AND status = 'pending'
`, workspaceID)
	// Not an error if there are no pending versions to discard.
	if errors.Is(err, internal.ErrResourceNotFound) {
		return nil
	}
	return err
}
