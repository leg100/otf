package api

import (
	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/organization"
)

func (m *jsonapiMarshaler) toOrganization(from *organization.Organization) *types.Organization {
	to := &types.Organization{
		Name:                       from.Name,
		CreatedAt:                  from.CreatedAt,
		ExternalID:                 from.ID,
		Permissions:                &types.DefaultOrganizationPermissions,
		SessionRemember:            from.SessionRemember,
		SessionTimeout:             from.SessionTimeout,
		AllowForceDeleteWorkspaces: from.AllowForceDeleteWorkspaces,
	}
	if from.Email != nil {
		to.Email = *from.Email
	}
	if from.CollaboratorAuthPolicy != nil {
		to.CollaboratorAuthPolicy = types.AuthPolicyType(*from.CollaboratorAuthPolicy)
	}
	return to
}

func (m *jsonapiMarshaler) toOrganizationList(from *organization.OrganizationList) (to []*types.Organization, opts []jsonapi.MarshalOption) {
	meta := jsonapi.MarshalMeta(map[string]*types.Pagination{
		"pagination": (*types.Pagination)(from.Pagination),
	})
	opts = append(opts, jsonapi.MarshalOption(meta))
	for _, item := range from.Items {
		to = append(to, m.toOrganization(item))
	}
	return
}
