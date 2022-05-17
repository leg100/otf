package otf

import (
	"encoding/json"
	"time"
)

type UserDBRow struct {
	UserID              string              `json:"user_id"`
	Username            string              `json:"username"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	CurrentOrganization *string             `json:"current_organization"`
	Sessions            []SessionDBRow      `json:"sessions"`
	Tokens              []TokenDBRow        `json:"tokens"`
	Organizations       []OrganizationDBRow `json:"organizations"`
}

func UnmarshalUserListFromDB(pgresult interface{}) (users []*User, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var rows []UserDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		ws, err := unmarshalUserDBRow(row)
		if err != nil {
			return nil, err
		}
		users = append(users, ws)
	}

	return users, nil
}

func UnmarshalUserFromDB(pgresult interface{}) (*User, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row UserDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}
	return unmarshalUserDBRow(row)
}

func unmarshalUserDBRow(row UserDBRow) (*User, error) {
	user := User{
		ID: row.UserID,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Username:            row.Username,
		CurrentOrganization: row.CurrentOrganization,
	}

	if row.Organizations != nil {
		orgs, err := UnmarshalOrganizationListFromDB(row.Organizations)
		if err != nil {
			return nil, err
		}
		user.Organizations = orgs
	}

	if row.Sessions != nil {
		sessions, err := UnmarshalSessionListFromDB(row.Sessions)
		if err != nil {
			return nil, err
		}
		user.Sessions = sessions
	}

	if row.Tokens != nil {
		tokens, err := UnmarshalTokenListFromDB(row.Tokens)
		if err != nil {
			return nil, err
		}
		user.Tokens = tokens
	}

	return &user, nil
}
