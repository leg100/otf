package state

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type (
	db interface {
		otf.DB

		createVersion(context.Context, *Version) error
		listVersions(ctx context.Context, opts StateVersionListOptions) (*VersionList, error)
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
		otf.DB // provides access to generated SQL queries
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
		return sql.Error(err)
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
	return db.Tx(ctx, func(tx otf.DB) error {
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
		Outputs:     make(outputList, len(row.StateVersionOutputs)),
	}
	for _, r := range row.StateVersionOutputs {
		sv.Outputs[r.Name.String] = outputRow(r).toOutput()
	}
	return &sv
}
