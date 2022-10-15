package otf

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

// UserResult represents the result of a database query for a user.
type UserResult struct {
	UserID        pgtype.Text           `json:"user_id"`
	Username      pgtype.Text           `json:"username"`
	CreatedAt     pgtype.Timestamptz    `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz    `json:"updated_at"`
	Sessions      []pggen.Sessions      `json:"sessions"`
	Tokens        []pggen.Tokens        `json:"tokens"`
	Organizations []pggen.Organizations `json:"organizations"`
	Teams         []pggen.Teams         `json:"teams"`
}

func UnmarshalUserResult(row UserResult, opts ...NewUserOption) (*User, error) {
	user := User{
		id:        row.UserID.String,
		createdAt: row.CreatedAt.Time.UTC(),
		updatedAt: row.UpdatedAt.Time.UTC(),
		username:  row.Username.String,
	}
	for _, or := range row.Organizations {
		user.Organizations = append(user.Organizations, UnmarshalOrganizationRow(or))
	}
	// Unmarshal team requires finding the team's corresponding
	// organization...pggen doesn't permit two layers of embedding table rows
	// (i.e. user -> team -> org)
	for _, tr := range row.Teams {
		for _, or := range row.Organizations {
			if tr.OrganizationID == or.OrganizationID {
				user.Teams = append(user.Teams, UnmarshalTeamResult(TeamResult{
					TeamID:                     tr.TeamID,
					Name:                       tr.Name,
					CreatedAt:                  tr.CreatedAt,
					OrganizationID:             tr.OrganizationID,
					PermissionManageWorkspaces: tr.PermissionManageWorkspaces,
					Organization:               &or,
				}))
			}
		}
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
	for _, o := range opts {
		o(&user)
	}

	return &user, nil
}
