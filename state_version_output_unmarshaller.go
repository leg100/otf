package otf

import (
	"encoding/json"
	"time"
)

type StateVersionOutputDBRow struct {
	StateVersionOutputID string    `json:"state_version_output_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Name                 string    `json:"name"`
	Sensitive            bool      `json:"sensitive"`
	Type                 string    `json:"type"`
	Value                string    `json:"value"`
	StateVersionID       string    `json:"state_version_id"`
}

func UnmarshalStateVersionOutputListFromDB(pgresult interface{}) (outputs []*StateVersionOutput, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var rows []StateVersionOutputDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		out, err := unmarshalStateVersionOutputDBRow(row)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, out)
	}

	return outputs, nil
}

func UnmarshalStateVersionOutputFromDB(pgresult interface{}) (*StateVersionOutput, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row StateVersionOutputDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}
	return unmarshalStateVersionOutputDBRow(row)
}

func unmarshalStateVersionOutputDBRow(row StateVersionOutputDBRow) (*StateVersionOutput, error) {
	out := StateVersionOutput{
		ID:        row.StateVersionOutputID,
		Sensitive: row.Sensitive,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Type:           row.Type,
		Value:          row.Value,
		Name:           row.Name,
		StateVersionID: row.StateVersionID,
	}

	return &out, nil
}
