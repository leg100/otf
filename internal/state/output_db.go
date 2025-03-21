package state

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type outputRow struct {
	StateVersionOutputID resource.TfeID `json:"state_version_output_id"`
	Name                 pgtype.Text    `json:"name"`
	Sensitive            pgtype.Bool    `json:"sensitive"`
	Type                 pgtype.Text    `json:"type"`
	Value                []byte         `json:"value"`
	StateVersionID       resource.TfeID `json:"state_version_id"`
}

// unmarshalVersionOutputRow unmarshals a database row into a state version
// output.
func (row outputRow) toOutput() *Output {
	return &Output{
		ID:             row.StateVersionOutputID,
		Sensitive:      row.Sensitive.Bool,
		Type:           row.Type.String,
		Value:          row.Value,
		Name:           row.Name.String,
		StateVersionID: row.StateVersionID,
	}
}

func (db *pgdb) getOutput(ctx context.Context, outputID resource.TfeID) (*Output, error) {
	result, err := q.FindStateVersionOutputByID(ctx, db.Conn(ctx), outputID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return outputRow(result).toOutput(), nil
}
