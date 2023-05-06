package api

import (
	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/workspace"
)

func (m *jsonapiMarshaler) toTags(from *workspace.TagList) (to []*types.OrganizationTag, opts []jsonapi.MarshalOption) {
	for _, ft := range from.Items {
		to = append(to, &types.OrganizationTag{
			ID:            ft.ID,
			Name:          ft.Name,
			InstanceCount: ft.InstanceCount,
			Organization: &types.Organization{
				Name: ft.Organization,
			},
		})
	}
	opts = []jsonapi.MarshalOption{toMarshalOption(from.Pagination)}
	return
}
