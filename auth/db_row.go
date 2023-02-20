package auth

import (
	"github.com/jackc/pgtype"
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

func (row userRow) toUser() *User {
	user := User{
		id:        row.UserID.String,
		createdAt: row.CreatedAt.Time.UTC(),
		updatedAt: row.UpdatedAt.Time.UTC(),
		username:  row.Username.String,
	}
	// avoid assigning empty slice to ensure equality in tests
	if len(row.Organizations) > 0 {
		user.organizations = row.Organizations
	}
	for _, tr := range row.Teams {
		user.teams = append(user.teams, teamRow(tr).toTeam())
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

func (row teamRow) toTeam() *Team {
	return &Team{
		id:           row.TeamID.String,
		createdAt:    row.CreatedAt.Time.UTC(),
		name:         row.Name.String,
		organization: row.OrganizationName.String,
		access: OrganizationAccess{
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

func (row agentTokenRow) toAgentToken() *AgentToken {
	return &AgentToken{
		id:           row.TokenID.String,
		createdAt:    row.CreatedAt.Time,
		token:        row.Token.String,
		description:  row.Description.String,
		organization: row.OrganizationName.String,
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
