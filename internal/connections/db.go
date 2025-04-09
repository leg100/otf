package connections

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type db struct {
	*sql.DB
}

func (db *db) createConnection(ctx context.Context, opts ConnectOptions) error {
	args := pgx.NamedArgs{
		"vcs_provider_id": opts.VCSProviderID,
		"repo_path":       sql.String(opts.RepoPath),
	}

	switch opts.ResourceID.Kind() {
	case resource.WorkspaceKind:
		args["workspace_id"] = &opts.ResourceID
	case resource.ModuleKind:
		args["module_id"] = &opts.ResourceID
	default:
		return fmt.Errorf("unsupported connection kind: %s", opts.ResourceID.Kind())
	}

	_, err := db.Exec(ctx, `
INSERT INTO repo_connections (
    vcs_provider_id,
    repo_path,
    workspace_id,
    module_id
) VALUES (
    @vcs_provider_id,
    @repo_path,
    @workspace_id,
    @module_id
)`, args)
	return err
}

func (db *db) deleteConnection(ctx context.Context, opts DisconnectOptions) (err error) {
	switch opts.ResourceID.Kind() {
	case resource.WorkspaceKind:
		_, err = db.Exec(ctx, `DELETE FROM repo_connections WHERE workspace_id = $1`, opts.ResourceID)
	case resource.ModuleKind:
		_, err = db.Exec(ctx, `DELETE FROM repo_connections WHERE module_id = $1`, opts.ResourceID)
	default:
		return fmt.Errorf("unknown connection kind: %v", opts.ResourceID)
	}
	return err
}
