package otf

import (
	"time"

	"github.com/leg100/otf/sql/pggen"
)

type UserDBResult struct {
	UserID              string                `json:"user_id"`
	Username            string                `json:"username"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	CurrentOrganization string                `json:"current_organization"`
	Sessions            []pggen.Sessions      `json:"sessions"`
	Tokens              []pggen.Tokens        `json:"tokens"`
	Organizations       []pggen.Organizations `json:"organizations"`
}

func UnmarshalUserDBResult(row UserDBResult) (*User, error) {
	user := User{
		id:        row.UserID,
		createdAt: row.CreatedAt,
		updatedAt: row.UpdatedAt,
		Username:  row.Username,
	}
	if row.CurrentOrganization != "" {
		user.CurrentOrganization = &row.CurrentOrganization
	}

	for _, typ := range row.Organizations {
		org, err := UnmarshalOrganizationDBResult(typ)
		if err != nil {
			return nil, err
		}
		user.Organizations = append(user.Organizations, org)
	}

	for _, typ := range row.Sessions {
		sess, err := UnmarshalSessionDBType(typ)
		if err != nil {
			return nil, err
		}
		user.Sessions = append(user.Sessions, sess)
	}

	for _, typ := range row.Tokens {
		token, err := unmarshalTokenDBType(typ)
		if err != nil {
			return nil, err
		}
		user.Tokens = append(user.Tokens, token)
	}

	return &user, nil
}
