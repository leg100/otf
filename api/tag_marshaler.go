package api

import (
	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/tags"
)

func (m *jsonapiMarshaler) toTags(from *tags.TagList) (to []*types.OrganizationTag, opts []jsonapi.MarshalOption) {
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
