package auth

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// userRow represents the result of a database query for a user.
type userRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []pggen.Teams      `json:"teams"`
}

func (row userRow) toUser() *otf.User {
	user := otf.User{
		ID:        row.UserID.String,
		CreatedAt: row.CreatedAt.Time.UTC(),
		UpdatedAt: row.UpdatedAt.Time.UTC(),
		Username:  row.Username.String,
	}
	// avoid assigning empty slice to ensure equality in tests
	if len(row.Organizations) > 0 {
		user.Organizations = row.Organizations
	}
	for _, tr := range row.Teams {
		user.Teams = append(user.Teams, teamRow(tr).toTeam())
	}

	return &user
}

// teamRow represents the result of a database query for a team.
type teamRow struct {
	TeamID                     pgtype.Text        `json:"team_id"`
	Name                       pgtype.Text        `json:"name"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	PermissionManageWorkspaces bool               `json:"permission_manage_workspaces"`
	PermissionManageVCS        bool               `json:"permission_manage_vcs"`
	PermissionManageRegistry   bool               `json:"permission_manage_registry"`
	OrganizationName           pgtype.Text        `json:"organization_name"`
}

func (row teamRow) toTeam() *otf.Team {
	return &otf.Team{
		ID:           row.TeamID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		Name:         row.Name.String,
		Organization: row.OrganizationName.String,
		Access: otf.OrganizationAccess{
			ManageWorkspaces: row.PermissionManageWorkspaces,
			ManageVCS:        row.PermissionManageVCS,
			ManageRegistry:   row.PermissionManageRegistry,
		},
	}
}

type agentTokenRow struct {
	TokenID          pgtype.Text        `json:"token_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Description      pgtype.Text        `json:"description"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

func (row agentTokenRow) toAgentToken() *otf.AgentToken {
	return &otf.AgentToken{
		ID:           row.TokenID.String,
		CreatedAt:    row.CreatedAt.Time,
		Token:        row.Token.String,
		Description:  row.Description.String,
		Organization: row.OrganizationName.String,
	}
}

type sessionRow struct {
	Token     pgtype.Text        `json:"token"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	Address   pgtype.Text        `json:"address"`
	Expiry    pgtype.Timestamptz `json:"expiry"`
	UserID    pgtype.Text        `json:"user_id"`
}

func (result sessionRow) toSession() *Session {
	return &Session{
		token:     result.Token.String,
		createdAt: result.CreatedAt.Time.UTC(),
		expiry:    result.Expiry.Time.UTC(),
		userID:    result.UserID.String,
		address:   result.Address.String,
	}
}

type registrySessionRow struct {
	Token            pgtype.Text        `json:"token"`
	Expiry           pgtype.Timestamptz `json:"expiry"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

func (result registrySessionRow) toRegistrySession() *registrySession {
	return &registrySession{
		token:        result.Token.String,
		expiry:       result.Expiry.Time.UTC(),
		organization: result.OrganizationName.String,
	}
}
