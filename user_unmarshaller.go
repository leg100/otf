package otf

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

// UserResult represents the result of a database query for a user.
type UserResult struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []pggen.Teams      `json:"teams"`
}

func UnmarshalUserResult(row UserResult, opts ...NewUserOption) *User {
	user := User{
		id:        row.UserID.String,
		createdAt: row.CreatedAt.Time.UTC(),
		updatedAt: row.UpdatedAt.Time.UTC(),
		username:  row.Username.String,
	}
	// avoid assigning empty slice to ensure equality in ./sql tests
	if len(row.Organizations) > 0 {
		user.organizations = row.Organizations
	}
	for _, tr := range row.Teams {
		user.teams = append(user.teams, UnmarshalTeamResult(TeamResult(tr)))
	}
	for _, o := range opts {
		o(&user)
	}

	return &user
}
