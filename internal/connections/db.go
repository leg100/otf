package connections

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type (
	db struct {
		*sql.DB
	}
)

func (db *db) createConnection(ctx context.Context, opts ConnectOptions) error {
	q := db.Querier(ctx)
	params := sqlc.InsertRepoConnectionParams{
		VCSProviderID: sql.ID(opts.VCSProviderID),
		RepoPath:      sql.String(opts.RepoPath),
	}

	switch opts.ConnectionType {
	case WorkspaceConnection:
		params.WorkspaceID = sql.ID(opts.ResourceID)
	case ModuleConnection:
		params.ModuleID = sql.ID(opts.ResourceID)
	default:
		return fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
	}

	if err := q.InsertRepoConnection(ctx, params); err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) deleteConnection(ctx context.Context, opts DisconnectOptions) (err error) {
	q := db.Querier(ctx)
	switch opts.ConnectionType {
	case WorkspaceConnection:
		_, err = q.DeleteWorkspaceConnectionByID(ctx, sql.ID(opts.ResourceID))
	case ModuleConnection:
		_, err = q.DeleteModuleConnectionByID(ctx, sql.ID(opts.ResourceID))
	default:
		return fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
	}
	return err
}
