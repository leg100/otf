package agent

import (
	"context"

	"github.com/leg100/otf/internal/sql"
)

type db struct {
	*sql.DB
}

func (db *db) createAgent(ctx context.Context, agent *Agent) error {
	//db.Conn(ctx).InsertAgent()
	return nil
}
