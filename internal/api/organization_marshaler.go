package api

import (
	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/organization"
)

func (m *jsonapiMarshaler) toOrganization(org *organization.Organization) *types.Organization {
	return &types.Organization{
		Name:            org.Name,
		CreatedAt:       org.CreatedAt,
		ExternalID:      org.ID,
		Permissions:     &types.DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	}
}

func (m *jsonapiMarshaler) toOrganizationList(from *organization.OrganizationList) (to []*types.Organization, opts []jsonapi.MarshalOption) {
	opts = []jsonapi.MarshalOption{toMarshalOption(from.Pagination)}
	for _, item := range from.Items {
		to = append(to, m.toOrganization(item))
	}
	return
}
