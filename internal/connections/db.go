package connections

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

type (
	db struct {
		*sql.DB
	}
)

func (db *db) createConnection(ctx context.Context, opts ConnectOptions) error {
	params := InsertRepoConnectionParams{
		VCSProviderID: opts.VCSProviderID,
		RepoPath:      sql.String(opts.RepoPath),
	}

	switch opts.ResourceID.Kind() {
	case resource.WorkspaceKind:
		params.WorkspaceID = &opts.ResourceID
	case resource.ModuleKind:
		params.ModuleID = &opts.ResourceID
	default:
		return fmt.Errorf("unsupported connection kind: %s", opts.ResourceID.Kind())
	}

	if err := q.InsertRepoConnection(ctx, db.Conn(ctx), params); err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) deleteConnection(ctx context.Context, opts DisconnectOptions) (err error) {
	switch opts.ResourceID.Kind() {
	case resource.WorkspaceKind:
		_, err = q.DeleteWorkspaceConnectionByID(ctx, db.Conn(ctx), &opts.ResourceID)
	case resource.ModuleKind:
		_, err = q.DeleteModuleConnectionByID(ctx, db.Conn(ctx), &opts.ResourceID)
	default:
		return fmt.Errorf("unknown connection kind: %v", opts.ResourceID)
	}
	return err
}
