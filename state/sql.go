package state

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is the hook database on postgres
type pgdb struct {
	otf.Database
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

// CreateStateVersion persists a StateVersion to the DB.
func (db *pgdb) CreateStateVersion(ctx context.Context, workspaceID string, sv *otf.StateVersion) error {
	return db.Transaction(ctx, func(tx otf.Database) error {
		_, err := tx.InsertStateVersion(ctx, pggen.InsertStateVersionParams{
			ID:          sql.String(sv.ID()),
			CreatedAt:   sql.Timestamptz(sv.CreatedAt()),
			Serial:      int(sv.Serial()),
			State:       sv.State(),
			WorkspaceID: sql.String(workspaceID),
		})
		if err != nil {
			return err
		}

		// Insert state_version_outputs
		for _, svo := range sv.Outputs() {
			_, err := tx.InsertStateVersionOutput(ctx, pggen.InsertStateVersionOutputParams{
				ID:             sql.String(svo.ID()),
				Name:           sql.String(svo.Name),
				Sensitive:      svo.Sensitive,
				Type:           sql.String(svo.Type),
				Value:          sql.String(svo.Value),
				StateVersionID: sql.String(sv.ID()),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) ListStateVersions(ctx context.Context, opts otf.StateVersionListOptions) (*StateVersionList, error) {
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

	var items []*StateVersion
	for _, r := range rows {
		sv, err := UnmarshalStateVersionResult(StateVersionResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, sv)
	}

	return &StateVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *pgdb) GetStateVersion(ctx context.Context, opts otf.StateVersionGetOptions) (*StateVersion, error) {
	if opts.ID != nil {
		result, err := db.FindStateVersionByID(ctx, sql.String(*opts.ID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return UnmarshalStateVersionResult(StateVersionResult(result))
	} else if opts.WorkspaceID != nil {
		result, err := db.FindStateVersionLatestByWorkspaceID(ctx, sql.String(*opts.WorkspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return UnmarshalStateVersionResult(StateVersionResult(result))
	} else {
		return nil, fmt.Errorf("no state version spec provided")
	}
}

func (db *pgdb) GetState(ctx context.Context, id string) ([]byte, error) {
	return db.FindStateVersionStateByID(ctx, sql.String(id))
}

// DeleteStateVersion deletes a state version from the DB
func (db *pgdb) DeleteStateVersion(ctx context.Context, id string) error {
	_, err := db.DeleteStateVersionByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
