package state

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type outputModel struct {
	StateVersionOutputID resource.TfeID `db:"state_version_output_id"`
	Name                 string         `db:"name"`
	Sensitive            bool           `db:"sensitive"`
	Type                 string         `db:"type"`
	Value                []byte         `db:"value"`
	StateVersionID       resource.TfeID `db:"state_version_id"`
}

func (m outputModel) toOutput() *Output {
	return &Output{
		ID:             m.StateVersionID,
		Name:           m.Name,
		Type:           m.Type,
		Value:          m.Value,
		Sensitive:      m.Sensitive,
		StateVersionID: m.StateVersionID,
	}
}

func (db *pgdb) getOutput(ctx context.Context, outputID resource.TfeID) (*Output, error) {
	rows := db.Query(ctx, `
SELECT state_version_output_id, name, sensitive, type, value, state_version_id
FROM state_version_outputs
WHERE state_version_output_id = $1
`, outputID)
	return sql.CollectOneRow(rows, scanOutput)
}

func scanOutput(row pgx.CollectableRow) (*Output, error) {
	model, err := pgx.RowToStructByName[outputModel](row)
	if err != nil {
		return nil, err
	}
	return model.toOutput(), nil
}
