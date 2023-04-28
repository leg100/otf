package api

import (
	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/organization"
)

func (m *jsonapiMarshaler) toOrganization(org *organization.Organization) *Organization {
	return &Organization{
		Name:            org.Name,
		CreatedAt:       org.CreatedAt,
		ExternalID:      org.ID,
		Permissions:     &DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	}
}

func (m *jsonapiMarshaler) toOrganizationList(from *organization.OrganizationList) (to []*Organization, opts []jsonapi.MarshalOption) {
	opts = []jsonapi.MarshalOption{jsonapi.MarshalMeta(NewPagination(from.Pagination))}
	for _, item := range from.Items {
		to = append(to, m.toOrganization(item))
	}
	return
}
