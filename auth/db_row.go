package auth

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

// userRow represents the result of a database query for a user.
type userRow struct {
	UserID    pgtype.Text        `json:"user_id"`
	Username  pgtype.Text        `json:"username"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	SiteAdmin bool               `json:"site_admin"`
	Teams     []pggen.Teams      `json:"teams"`
}

func (row userRow) toUser() *User {
	user := User{
		ID:        row.UserID.String,
		CreatedAt: row.CreatedAt.Time.UTC(),
		UpdatedAt: row.UpdatedAt.Time.UTC(),
		Username:  row.Username.String,
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

func (row teamRow) toTeam() *Team {
	return &Team{
		ID:           row.TeamID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		Name:         row.Name.String,
		Organization: row.OrganizationName.String,
		Access: OrganizationAccess{
			ManageWorkspaces: row.PermissionManageWorkspaces,
			ManageVCS:        row.PermissionManageVCS,
			ManageRegistry:   row.PermissionManageRegistry,
		},
	}
}
