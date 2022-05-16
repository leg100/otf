package otf

import (
	"encoding/json"
	"time"
)

type SessionDBRow struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Address   string    `json:"address"`
	Flash     []byte    `json:"flash"`
	Expiry    time.Time `json:"expiry"`
	UserID    string    `json:"user_id"`
}

func UnmarshalSessionListFromDB(pgresult interface{}) (sessions []*Session, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var rows []SessionDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		session, err := unmarshalSessionDBRow(row)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func UnmarshalSessionFromDB(pgresult interface{}) (*Session, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row SessionDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}
	return unmarshalSessionDBRow(row)
}

func unmarshalSessionDBRow(row SessionDBRow) (*Session, error) {
	session := Session{
		Token: row.Token,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Expiry: row.Expiry,
		UserID: row.UserID,
		SessionData: SessionData{
			Address: row.Address,
		},
	}

	return &session, nil
}
