package otf

import (
	"encoding/json"
	"time"
)

type TokenDBRow struct {
	TokenID     string    `json:"token_id"`
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
}

func UnmarshalTokenListFromDB(pgresult interface{}) (tokens []*Token, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var rows []TokenDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		token, err := unmarshalTokenDBRow(row)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func UnmarshalTokenFromDB(pgresult interface{}) (*Token, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row TokenDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}
	return unmarshalTokenDBRow(row)
}

func unmarshalTokenDBRow(row TokenDBRow) (*Token, error) {
	token := Token{
		ID: row.UserID,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Token:       row.TokenID,
		Description: row.Description,
		UserID:      row.UserID,
	}

	return &token, nil
}
