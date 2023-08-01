package workspace

import (
	"context"

	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

func (db *pgdb) listRemoteStateConsumers(ctx context.Context, workspaceID string) ([]*Workspace, error) {
	q := db.Conn(ctx)
	rows, err := q.FindRemoteStateConsumersByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	var items []*Workspace
	for _, r := range rows {
		ws, err := pgresult(r).toWorkspace()
		if err != nil {
			return nil, err
		}
		items = append(items, ws)
	}
	return items, nil
}

func (db *pgdb) replaceRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error {
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		// delete any existing consumers first
		if err := db.deleteRemoteStateConsumers(ctx, workspaceID, consumers); err != nil {
			return err
		}
		for _, consumer := range consumers {
			_, err := q.InsertRemoteStateConsumerByID(ctx, sql.String(workspaceID), sql.String(consumer))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) addRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error {
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		for _, consumer := range consumers {
			_, err := q.InsertRemoteStateConsumerByID(ctx, sql.String(workspaceID), sql.String(consumer))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error {
	q := db.Conn(ctx)
	_, err := q.DeleteRemoteStateConsumersByConsumerIDs(ctx, sql.String(workspaceID), consumers)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
