package api

import (
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
		CostEstimationEnabled:      from.CostEstimationEnabled,
	}
	if from.Email != nil {
		to.Email = *from.Email
	}
	if from.CollaboratorAuthPolicy != nil {
		to.CollaboratorAuthPolicy = types.AuthPolicyType(*from.CollaboratorAuthPolicy)
	}
	return to
}
