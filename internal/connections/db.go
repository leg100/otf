package connections

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/resource"
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
		VCSProviderID: opts.VCSProviderID,
		RepoPath:      sql.String(opts.RepoPath),
	}

	switch opts.ResourceID.Kind {
	case resource.WorkspaceKind:
		params.WorkspaceID = &opts.ResourceID
	case resource.ModuleKind:
		params.ModuleID = &opts.ResourceID
	default:
		return fmt.Errorf("unsupported connection kind: %s", opts.ResourceID.Kind)
	}

	if err := q.InsertRepoConnection(ctx, params); err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) deleteConnection(ctx context.Context, opts DisconnectOptions) (err error) {
	q := db.Querier(ctx)
	switch opts.ResourceID.Kind {
	case resource.WorkspaceKind:
		_, err = q.DeleteWorkspaceConnectionByID(ctx, &opts.ResourceID)
	case resource.ModuleKind:
		_, err = q.DeleteModuleConnectionByID(ctx, &opts.ResourceID)
	default:
		return fmt.Errorf("unknown connection kind: %v", opts.ResourceID)
	}
	return err
}
