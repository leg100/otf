package state

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql"
)

type outputRow struct {
	StateVersionOutputID pgtype.Text `json:"state_version_output_id"`
	Name                 pgtype.Text `json:"name"`
	Sensitive            bool        `json:"sensitive"`
	Type                 pgtype.Text `json:"type"`
	Value                pgtype.Text `json:"value"`
	StateVersionID       pgtype.Text `json:"state_version_id"`
}

func (db *pgdb) getOutput(ctx context.Context, outputID string) (*Output, error) {
	result, err := db.FindStateVersionOutputByID(ctx, sql.String(outputID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return outputRow(result).toOutput(), nil
}

// unmarshalVersionOutputRow unmarshals a database row into a state version
// output.
func (row outputRow) toOutput() *Output {
	return &Output{
		ID:             row.StateVersionOutputID.String,
		Sensitive:      row.Sensitive,
		Type:            row.Type.String,
		Value:          row.Value.String,
		Name:           row.Name.String,
		StateVersionID: row.StateVersionID.String,
	}
}
