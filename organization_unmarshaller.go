package otf

import (
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
