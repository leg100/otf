package sql

import (
	"time"

	"github.com/leg100/otf"
)

type userRow struct {
	UserID              *string         `json:"user_id"`
	Username            *string         `json:"username"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	CurrentOrganization *string         `json:"current_organization"`
	Sessions            []Sessions      `json:"sessions"`
	Tokens              []Tokens        `json:"tokens"`
	Organizations       []Organizations `json:"organizations"`
}

func convertUserRow(row userRow) *otf.User {
	user := otf.User{}
	user.ID = *row.UserID
	user.Username = *row.Username
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	user.CurrentOrganization = row.CurrentOrganization

	for _, org := range row.Organizations {
		user.Organizations = append(user.Organizations, convertOrganizationComposite(org))
	}
	for _, token := range row.Tokens {
		user.Tokens = append(user.Tokens, convertTokenComposite(token))
	}
	for _, org := range row.Sessions {
		user.Sessions = append(user.Sessions, convertSessionComposite(org))
	}
	return &user
}
