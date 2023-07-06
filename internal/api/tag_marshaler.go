package api

import (
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/workspace"
)

func (m *jsonapiMarshaler) toTag(from *workspace.Tag) *types.OrganizationTag {
	return &types.OrganizationTag{
		ID:            from.ID,
		Name:          from.Name,
		InstanceCount: from.InstanceCount,
		Organization: &types.Organization{
			Name: from.Organization,
		},
	}
}
