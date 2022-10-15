package otf

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

// TeamResult represents the result of a database query for a team.
type TeamResult struct {
	TeamID                     pgtype.Text          `json:"team_id"`
	Name                       pgtype.Text          `json:"name"`
	CreatedAt                  pgtype.Timestamptz   `json:"created_at"`
	OrganizationID             pgtype.Text          `json:"organization_id"`
	PermissionManageWorkspaces bool                 `json:"permission_manage_workspaces"`
	Organization               *pggen.Organizations `json:"organization"`
}

func UnmarshalTeamResult(row TeamResult, opts ...NewTeamOption) *Team {
	return &Team{
		id:           row.TeamID.String,
		createdAt:    row.CreatedAt.Time.UTC(),
		name:         row.Name.String,
		organization: UnmarshalOrganizationRow(*row.Organization),
		access: OrganizationAccess{
			ManageWorkspaces: row.PermissionManageWorkspaces,
		},
	}
}
