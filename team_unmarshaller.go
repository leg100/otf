package otf

import (
	"github.com/jackc/pgtype"
)

// TeamDBResult is a row from the database teams table
//
// TODO: rename TeamDBResult to TeamRow
type TeamDBResult struct {
	TeamID                     pgtype.Text        `json:"team_id"`
	Name                       pgtype.Text        `json:"name"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	OrganizationID             pgtype.Text        `json:"organization_id"`
	PermissionManageWorkspaces bool               `json:"permission_manage_workspaces"`
	OrganizationName           pgtype.Text        `json:"organization_name"`
}

func UnmarshalTeamDBResult(row TeamDBResult, opts ...NewTeamOption) *Team {
	return &Team{
		id:               row.TeamID.String,
		createdAt:        row.CreatedAt.Time.UTC(),
		name:             row.Name.String,
		organizationName: row.OrganizationName.String,
		organizationID:   row.OrganizationID.String,
		access: OrganizationAccess{
			ManageWorkspaces: row.PermissionManageWorkspaces,
		},
	}
}
