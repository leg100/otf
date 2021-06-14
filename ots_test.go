package ots

import (
	"testing"

	"github.com/google/jsonapi"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name      string
		opts      *tfe.ListOptions
		count     int
		want      *Pagination
		wantLinks *jsonapi.Links
		wantMeta  *jsonapi.Meta
	}{
		{
			name: "one page",
			opts: &tfe.ListOptions{
				PageNumber: 1,
				PageSize:   20,
			},
			count: 5,
			want: &Pagination{
				CurrentPage: 1,
				TotalPages:  1,
				TotalCount:  5,
				PageSize:    20,
				path:        "/v2/api/foobar",
			},
			wantLinks: &jsonapi.Links{
				"first": "/v2/api/foobar?page%5Bnumber%5D=1&page%5Bsize%5D=20",
				"last":  "/v2/api/foobar?page%5Bnumber%5D=1&page%5Bsize%5D=20",
				"self":  "/v2/api/foobar?page%5Bnumber%5D=1&page%5Bsize%5D=20",
			},
			wantMeta: &jsonapi.Meta{
				"pagination": map[string]interface{}{
					"current-page": 1,
					"total-count":  5,
					"total-pages":  1,
				},
			},
		},
		{
			name: "multiple pages",
			opts: &tfe.ListOptions{
				PageNumber: 3,
				PageSize:   20,
			},
			count: 101,
			want: &Pagination{
				CurrentPage:  3,
				PreviousPage: Int(2),
				NextPage:     Int(4),
				TotalPages:   6,
				TotalCount:   101,
				PageSize:     20,
				path:         "/v2/api/foobar",
			},
			wantLinks: &jsonapi.Links{
				"first": "/v2/api/foobar?page%5Bnumber%5D=1&page%5Bsize%5D=20",
				"last":  "/v2/api/foobar?page%5Bnumber%5D=6&page%5Bsize%5D=20",
				"prev":  "/v2/api/foobar?page%5Bnumber%5D=2&page%5Bsize%5D=20",
				"next":  "/v2/api/foobar?page%5Bnumber%5D=4&page%5Bsize%5D=20",
				"self":  "/v2/api/foobar?page%5Bnumber%5D=3&page%5Bsize%5D=20",
			},
			wantMeta: &jsonapi.Meta{
				"pagination": map[string]interface{}{
					"current-page": 3,
					"prev-page":    Int(2),
					"next-page":    Int(4),
					"total-count":  101,
					"total-pages":  6,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPagination("/v2/api/foobar", tt.opts, tt.count)
			assert.Equal(t, tt.want, p)
			assert.Equal(t, tt.wantLinks, p.JSONAPIPaginationLinks())
			assert.Equal(t, tt.wantMeta, p.JSONAPIPaginationMeta())
		})
	}
}
