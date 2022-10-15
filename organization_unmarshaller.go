package otf

import (
	"github.com/leg100/otf/http/dto"
	"github.com/leg100/otf/sql/pggen"
)

func UnmarshalOrganizationDBResult(result pggen.Organizations) *Organization {
	return &Organization{
		id:              result.OrganizationID.String,
		createdAt:       result.CreatedAt.Time.UTC(),
		updatedAt:       result.UpdatedAt.Time.UTC(),
		name:            result.Name.String,
		sessionRemember: result.SessionRemember,
		sessionTimeout:  result.SessionTimeout,
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
