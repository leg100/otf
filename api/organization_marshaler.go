package api

import (
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/organization"
)

func (m *jsonapiMarshaler) toOrganization(org *organization.Organization) *jsonapi.Organization {
	return &jsonapi.Organization{
		Name:            org.Name,
		CreatedAt:       org.CreatedAt,
		ExternalID:      org.ID,
		Permissions:     &jsonapi.DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	}
}

func (m *jsonapiMarshaler) toOrganizationList(from *organization.OrganizationList) *jsonapi.OrganizationList {
	to := &jsonapi.OrganizationList{
		Pagination: jsonapi.NewPagination(from.Pagination),
	}
	for _, item := range from.Items {
		to.Items = append(to.Items, m.toOrganization(item))
	}
	return to
}
