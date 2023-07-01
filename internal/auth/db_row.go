package auth

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal/sql/pggen"
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
		SiteAdmin: row.SiteAdmin,
	}
	for _, tr := range row.Teams {
		user.Teams = append(user.Teams, teamRow(tr).toTeam())
	}

	return &user
}

// teamRow represents the result of a database query for a team.
type teamRow struct {
	TeamID                          pgtype.Text        `json:"team_id"`
	Name                            pgtype.Text        `json:"name"`
	CreatedAt                       pgtype.Timestamptz `json:"created_at"`
	PermissionManageWorkspaces      bool               `json:"permission_manage_workspaces"`
	PermissionManageVCS             bool               `json:"permission_manage_vcs"`
	PermissionManageModules         bool               `json:"permission_manage_modules"`
	OrganizationName                pgtype.Text        `json:"organization_name"`
	SSOTeamID                       pgtype.Text        `json:"sso_team_id"`
	Visibility                      pgtype.Text        `json:"visibility"`
	PermissionManagePolicies        bool               `json:"permission_manage_policies"`
	PermissionManagePolicyOverrides bool               `json:"permission_manage_policy_overrides"`
	PermissionManageProviders       bool               `json:"permission_manage_providers"`
}

func (row teamRow) toTeam() *Team {
	to := Team{
		ID:           row.TeamID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		Name:         row.Name.String,
		Organization: row.OrganizationName.String,
		Visibility:   row.Visibility.String,
		Access: OrganizationAccess{
			ManageWorkspaces:      row.PermissionManageWorkspaces,
			ManageVCS:             row.PermissionManageVCS,
			ManageModules:         row.PermissionManageModules,
			ManageProviders:       row.PermissionManageProviders,
			ManagePolicies:        row.PermissionManagePolicies,
			ManagePolicyOverrides: row.PermissionManagePolicyOverrides,
		},
	}
	if row.SSOTeamID.Status == pgtype.Present {
		to.SSOTeamID = &row.SSOTeamID.String
	}
	return &to
}
