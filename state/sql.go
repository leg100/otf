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

	createVersion(context.Context, *Version) error
	listVersions(ctx context.Context, opts StateVersionListOptions) (*VersionList, error)
	getVersion(ctx context.Context, opts StateVersionGetOptions) (*Version, error)
	getState(ctx context.Context, versionID string) ([]byte, error)
	deleteVersion(ctx context.Context, versionID string) error
}

// pgdb is a state/state-version database on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

func (db *pgdb) createVersion(ctx context.Context, v *Version) error {
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

		// Insert state_version_outputs
		for _, svo := range v.Outputs() {
			_, err := db.InsertStateVersionOutput(ctx, pggen.InsertStateVersionOutputParams{
				ID:             sql.String(svo.id),
				Name:           sql.String(svo.Name),
				Sensitive:      svo.Sensitive,
				Type:           sql.String(svo.Type),
				Value:          sql.String(svo.Value),
				StateVersionID: sql.String(v.id),
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
		sv, err := unmarshalVersionRow(versionRow(r))
		if err != nil {
			return nil, err
		}
		items = append(items, sv)
	}

	return &VersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *pgdb) getVersion(ctx context.Context, opts StateVersionGetOptions) (*Version, error) {
	if opts.ID != nil {
		result, err := db.FindStateVersionByID(ctx, sql.String(*opts.ID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return unmarshalVersionRow(versionRow(result))
	} else if opts.WorkspaceID != nil {
		result, err := db.FindStateVersionLatestByWorkspaceID(ctx, sql.String(*opts.WorkspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return unmarshalVersionRow(versionRow(result))
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

// versionRow represents the result of a database query for a state
// version.
type versionRow struct {
	StateVersionID      pgtype.Text                 `json:"state_version_id"`
	CreatedAt           pgtype.Timestamptz          `json:"created_at"`
	Serial              int                         `json:"serial"`
	State               []byte                      `json:"state"`
	WorkspaceID         pgtype.Text                 `json:"workspace_id"`
	StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
}

// unmarshalVersionRow unmarshals a database row into a state version.
func unmarshalVersionRow(row versionRow) (*Version, error) {
	sv := Version{
		id:          row.StateVersionID.String,
		createdAt:   row.CreatedAt.Time.UTC(),
		serial:      int64(row.Serial),
		state:       row.State,
		workspaceID: row.WorkspaceID.String,
	}
	for _, r := range row.StateVersionOutputs {
		sv.outputs = append(sv.outputs, UnmarshalStateVersionOutputRow(r))
	}
	return &sv, nil
}
