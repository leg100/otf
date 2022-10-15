package otf

import (
	"github.com/leg100/otf/http/dto"
	"github.com/leg100/otf/sql/pggen"
)

// UnmarshalOrganizationRow converts an organization database row into an
// organization.
func UnmarshalOrganizationRow(row pggen.Organizations) *Organization {
	return &Organization{
		id:              row.OrganizationID.String,
		createdAt:       row.CreatedAt.Time.UTC(),
		updatedAt:       row.UpdatedAt.Time.UTC(),
		name:            row.Name.String,
		sessionRemember: row.SessionRemember,
		sessionTimeout:  row.SessionTimeout,
	}
}

func UnmarshalOrganizationJSONAPI(model *dto.Organization) *Organization {
	return &Organization{
		id:              model.ExternalID,
		createdAt:       model.CreatedAt,
		name:            model.Name,
		sessionRemember: model.SessionRemember,
		sessionTimeout:  model.SessionTimeout,
	}
}
