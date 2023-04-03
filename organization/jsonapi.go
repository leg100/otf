package organization

import (
	"github.com/leg100/otf/http/jsonapi"
)

// JSONAPIMarshaler marshals workspace into a struct suitable for marshaling
// into json-api
type JSONAPIMarshaler struct{}

func (m *JSONAPIMarshaler) ToOrganization(org *Organization) *jsonapi.Organization {
	return &jsonapi.Organization{
		Name:            org.Name,
		CreatedAt:       org.CreatedAt,
		ExternalID:      org.ID,
		Permissions:     &jsonapi.DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	}
}

func (m *JSONAPIMarshaler) toList(from *OrganizationList) *jsonapi.OrganizationList {
	to := &jsonapi.OrganizationList{
		Pagination: jsonapi.NewPagination(from.Pagination),
	}
	for _, item := range from.Items {
		to.Items = append(to.Items, m.ToOrganization(item))
	}
	return to
}
