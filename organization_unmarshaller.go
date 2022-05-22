package otf

import (
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/sql/pggen"
)

func UnmarshalOrganizationDBResult(result pggen.Organizations) (*Organization, error) {
	org := Organization{
		ID: result.OrganizationID,
		Timestamps: Timestamps{
			CreatedAt: result.CreatedAt.Local(),
			UpdatedAt: result.UpdatedAt.Local(),
		},
		Name:            result.Name,
		SessionRemember: result.SessionRemember,
		SessionTimeout:  result.SessionTimeout,
	}

	return &org, nil
}

func UmarshalOrganizationJSONAPI(model *jsonapi.Organization) *Organization {
	return &Organization{
		ID:              model.ExternalID,
		name:            model.Name,
		sessionRemember: model.SessionRemember,
		sessionTimeout:  model.SessionTimeout,
	}
}
