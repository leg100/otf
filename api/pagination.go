package api

import "github.com/leg100/otf"

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	CurrentPage  int  `json:"current-page"`
	PreviousPage *int `json:"prev-page"`
	NextPage     *int `json:"next-page"`
	TotalPages   int  `json:"total-pages"`
	TotalCount   int  `json:"total-count"`
}

func NewPagination(p *otf.Pagination) *Pagination {
	return &Pagination{
		CurrentPage:  p.Opts.SanitizedPageNumber(),
		PreviousPage: p.PrevPage(),
		NextPage:     p.NextPage(),
		TotalPages:   p.TotalPages(),
		TotalCount:   p.Count,
	}
}

// NewPaginationFromJSONAPI constructs pagination from a json:api struct
func NewPaginationFromJSONAPI(json *Pagination) *otf.Pagination {
	return &otf.Pagination{
		Count: json.TotalCount,
		// we can't determine the page size so we'll just pass in 0 which
		// ListOptions interprets as the default page size
		Opts: otf.ListOptions{PageNumber: json.CurrentPage, PageSize: 0},
	}
}
