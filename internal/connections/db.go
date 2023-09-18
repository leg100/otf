package connections

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	db struct {
		*sql.DB
	}
)

func (db *db) createConnection(ctx context.Context, hookID uuid.UUID, opts ConnectOptions) error {
	q := db.Conn(ctx)
	params := pggen.InsertRepoConnectionParams{
		WebhookID: sql.UUID(hookID),
	}

	switch opts.ConnectionType {
	case WorkspaceConnection:
		params.WorkspaceID = sql.String(opts.ResourceID)
		params.ModuleID = pgtype.Text{Status: pgtype.Null}
	case ModuleConnection:
		params.ModuleID = sql.String(opts.ResourceID)
		params.WorkspaceID = pgtype.Text{Status: pgtype.Null}
	default:
		return fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
	}

	if _, err := q.InsertRepoConnection(ctx, params); err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) deleteConnection(ctx context.Context, opts DisconnectOptions) (err error) {
	q := db.Conn(ctx)
	switch opts.ConnectionType {
	case WorkspaceConnection:
		_, err = q.DeleteWorkspaceConnectionByID(ctx, sql.String(opts.ResourceID))
	case ModuleConnection:
		_, err = q.DeleteModuleConnectionByID(ctx, sql.String(opts.ResourceID))
	default:
		return fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
	}
	return err
}
