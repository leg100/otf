package api

import (
	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/api/types"
)

func toMarshalOption(p *otf.Pagination) jsonapi.MarshalOption {
	return jsonapi.MarshalMeta(map[string]any{"pagination": &types.Pagination{
		CurrentPage:  p.Opts.SanitizedPageNumber(),
		PreviousPage: p.PrevPage(),
		NextPage:     p.NextPage(),
		TotalPages:   p.TotalPages(),
		TotalCount:   p.Count,
	}})
}
