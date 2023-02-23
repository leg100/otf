package state

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a database of state and state versions
type db interface {
	otf.Database

	createVersion(context.Context, *version) error
	listVersions(ctx context.Context, opts stateVersionListOptions) (*versionList, error)
	getVersion(ctx context.Context, opts stateVersionGetOptions) (*version, error)
	getState(ctx context.Context, versionID string) ([]byte, error)
	deleteVersion(ctx context.Context, versionID string) error
	getOutput(ctx context.Context, outputID string) (*output, error)
}

// pgdb is a state/state-version database on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

func (db *pgdb) createVersion(ctx context.Context, v *version) error {
	return db.Transaction(ctx, func(db otf.Database) error {
		_, err := db.InsertStateVersion(ctx, pggen.InsertStateVersionParams{
			ID:          sql.String(v.id),
			CreatedAt:   sql.Timestamptz(v.createdAt),
			Serial:      int(v.serial),
			State:       v.state,
			WorkspaceID: sql.String(v.workspaceID),
		})
		if err != nil {
			return err
		}

		for _, svo := range v.outputs {
			_, err := db.InsertStateVersionOutput(ctx, pggen.InsertStateVersionOutputParams{
				ID:             sql.String(svo.id),
				Name:           sql.String(svo.name),
				Sensitive:      svo.sensitive,
				Type:           sql.String(svo.typ),
				Value:          sql.String(svo.value),
				StateVersionID: sql.String(v.id),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) listVersions(ctx context.Context, opts stateVersionListOptions) (*versionList, error) {
	batch := &pgx.Batch{}

	db.FindStateVersionsByWorkspaceNameBatch(batch, pggen.FindStateVersionsByWorkspaceNameParams{
		WorkspaceName:    sql.String(opts.Workspace),
		OrganizationName: sql.String(opts.Organization),
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	db.CountStateVersionsByWorkspaceNameBatch(batch, sql.String(opts.Workspace), sql.String(opts.Organization))

	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindStateVersionsByWorkspaceNameScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountStateVersionsByWorkspaceNameScan(results)
	if err != nil {
		return nil, err
	}

	var items []*version
	for _, r := range rows {
		items = append(items, pgRow(r).toVersion())
	}

	return &versionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *pgdb) getVersion(ctx context.Context, opts stateVersionGetOptions) (*version, error) {
	if opts.ID != nil {
		result, err := db.FindStateVersionByID(ctx, sql.String(*opts.ID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toVersion(), nil
	} else if opts.WorkspaceID != nil {
		result, err := db.FindStateVersionLatestByWorkspaceID(ctx, sql.String(*opts.WorkspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toVersion(), nil
	} else {
		return nil, fmt.Errorf("no state version spec provided")
	}
}

func (db *pgdb) getState(ctx context.Context, id string) ([]byte, error) {
	return db.FindStateVersionStateByID(ctx, sql.String(id))
}

// deleteVersion deletes a state version from the DB
func (db *pgdb) deleteVersion(ctx context.Context, id string) error {
	_, err := db.DeleteStateVersionByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// pgRow is a row from a postgres query for a state version.
type pgRow struct {
	StateVersionID      pgtype.Text                 `json:"state_version_id"`
	CreatedAt           pgtype.Timestamptz          `json:"created_at"`
	Serial              int                         `json:"serial"`
	State               []byte                      `json:"state"`
	WorkspaceID         pgtype.Text                 `json:"workspace_id"`
	StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
}

func (row pgRow) toVersion() *version {
	sv := version{
		id:          row.StateVersionID.String,
		createdAt:   row.CreatedAt.Time.UTC(),
		serial:      int64(row.Serial),
		state:       row.State,
		workspaceID: row.WorkspaceID.String,
		outputs:     make(outputList, len(row.StateVersionOutputs)),
	}
	for _, r := range row.StateVersionOutputs {
		sv.outputs[r.Name.String] = outputRow(r).toOutput()
	}
	return &sv
}
