package otf

import (
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

type UserDBResult struct {
	UserID        pgtype.Text           `json:"user_id"`
	Username      pgtype.Text           `json:"username"`
	CreatedAt     pgtype.Timestamptz    `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz    `json:"updated_at"`
	Sessions      []pggen.Sessions      `json:"sessions"`
	Tokens        []pggen.Tokens        `json:"tokens"`
	Organizations []pggen.Organizations `json:"organizations"`
	Teams         []pggen.Teams         `json:"teams"`
}

func UnmarshalUserDBResult(row UserDBResult, opts ...NewUserOption) (*User, error) {
	user := User{
		id:        row.UserID.String,
		createdAt: row.CreatedAt.Time.UTC(),
		updatedAt: row.UpdatedAt.Time.UTC(),
		username:  row.Username.String,
	}
	// build a mapping of organization id to name whilst reconstructing
	// organizations...
	organizationIDNameMap := make(map[string]string)
	for _, or := range row.Organizations {
		org, err := UnmarshalOrganizationDBResult(or)
		if err != nil {
			return nil, err
		}
		user.Organizations = append(user.Organizations, org)
		organizationIDNameMap[org.ID()] = org.Name()
	}
	// ...reconstruct teams and use the organization mapping to retrieve the
	// organization name. (We do this here to avoid an overly nested SQL
	// query).
	for _, tr := range row.Teams {
		orgName, ok := organizationIDNameMap[tr.OrganizationID.String]
		if !ok {
			return nil, fmt.Errorf("constructing user teams: no name maps to organization ID: %s", tr.OrganizationID.String)
		}
		user.Teams = append(user.Teams, UnmarshalTeamDBResult(TeamDBResult{
			TeamID:           tr.TeamID,
			Name:             tr.Name,
			CreatedAt:        tr.CreatedAt,
			OrganizationID:   tr.OrganizationID,
			OrganizationName: pgtype.Text{String: orgName, Status: pgtype.Present},
		}))
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
