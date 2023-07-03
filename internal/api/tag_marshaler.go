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
	meta := jsonapi.MarshalMeta(map[string]*types.Pagination{
		"pagination": (*types.Pagination)(from.Pagination),
	})
	opts = append(opts, jsonapi.MarshalOption(meta))
	return
}
