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
	otf.DB

	createVersion(context.Context, *Version) error
	listVersions(ctx context.Context, opts StateVersionListOptions) (*VersionList, error)
	getVersion(ctx context.Context, opts stateVersionGetOptions) (*Version, error)
	getState(ctx context.Context, versionID string) ([]byte, error)
	deleteVersion(ctx context.Context, versionID string) error
	getOutput(ctx context.Context, outputID string) (*Output, error)
}

// pgdb is a state/state-version database on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
}

func newPGDB(db otf.DB) *pgdb {
	return &pgdb{db}
}

func (db *pgdb) createVersion(ctx context.Context, v *Version) error {
	return db.Tx(ctx, func(db otf.DB) error {
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

func (db *pgdb) listVersions(ctx context.Context, opts StateVersionListOptions) (*VersionList, error) {
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

	var items []*Version
	for _, r := range rows {
		items = append(items, pgRow(r).toVersion())
	}

	return &VersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *pgdb) getVersion(ctx context.Context, opts stateVersionGetOptions) (*Version, error) {
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

func (row pgRow) toVersion() *Version {
	sv := Version{
		ID:          row.StateVersionID.String,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Serial:      int64(row.Serial),
		State:       row.State,
		WorkspaceID: row.WorkspaceID.String,
		Outputs:     make(outputList, len(row.StateVersionOutputs)),
	}
	for _, r := range row.StateVersionOutputs {
		sv.Outputs[r.Name.String] = outputRow(r).toOutput()
	}
	return &sv
}
