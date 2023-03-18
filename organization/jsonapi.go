package organization

import (
	"github.com/leg100/otf/http/jsonapi"
)

// jsonapiMarshaler marshals workspace into a struct suitable for marshaling
// into json-api
type jsonapiMarshaler struct{}

func (m *jsonapiMarshaler) toOrganization(org *Organization) *jsonapi.Organization {
	return &jsonapi.Organization{
		Name:            org.Name,
		CreatedAt:       org.CreatedAt,
		ExternalID:      org.ID,
		Permissions:     &jsonapi.DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	}
}

func (m *jsonapiMarshaler) toList(from *OrganizationList) *jsonapi.OrganizationList {
	to := &jsonapi.OrganizationList{
		Pagination: from.Pagination.ToJSONAPI(),
	}
	for _, item := range from.Items {
		to.Items = append(to.Items, m.toOrganization(item))
	}
	return to
}

func newFromJSONAPI(from jsonapi.Organization) *Organization {
	return &Organization{
		ID:              from.ExternalID,
		CreatedAt:       from.CreatedAt,
		Name:            from.Name,
		SessionRemember: from.SessionRemember,
		SessionTimeout:  from.SessionTimeout,
	}
}
