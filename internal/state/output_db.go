package state

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type outputRow struct {
	StateVersionOutputID resource.TfeID `db:"state_version_output_id"`
	Name                 string         `db:"name"`
	Sensitive            bool           `db:"sensitive"`
	Type                 string         `db:"type"`
	Value                []byte         `db:"value"`
	StateVersionID       resource.TfeID `db:"state_version_id"`
}

func (db *pgdb) getOutput(ctx context.Context, outputID resource.TfeID) (*Output, error) {
	rows := db.Query(ctx, `
SELECT state_version_output_id, name, sensitive, type, value, state_version_id
FROM state_version_outputs
WHERE state_version_output_id = $1
`, outputID)
	return sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[agentToken])
}
