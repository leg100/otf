package otf

import (
	"github.com/leg100/otf/http/dto"
	"github.com/leg100/otf/sql/pggen"
)

func UnmarshalOrganizationDBResult(result pggen.Organizations) (*Organization, error) {
	org := Organization{
		id:              result.OrganizationID,
		createdAt:       result.CreatedAt,
		updatedAt:       result.UpdatedAt,
		name:            result.Name,
		sessionRemember: result.SessionRemember,
		sessionTimeout:  result.SessionTimeout,
	}

	return &org, nil
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
