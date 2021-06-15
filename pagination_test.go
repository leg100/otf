package ots

import (
	"testing"

	"github.com/google/jsonapi"
	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name      string
		paginated Paginated
		wantLinks *jsonapi.Links
		wantMeta  *jsonapi.Meta
	}{
		{
			name:      "one page",
			paginated: NewPaginatedMock("/v2/api/foobar", 5, 1, 20),
			wantLinks: &jsonapi.Links{
				"first": "/v2/api/foobar?page%5Bnumber%5D=1&page%5Bsize%5D=20",
				"last":  "/v2/api/foobar?page%5Bnumber%5D=1&page%5Bsize%5D=20",
				"self":  "/v2/api/foobar?page%5Bnumber%5D=1&page%5Bsize%5D=20",
			},
			wantMeta: &jsonapi.Meta{
				"pagination": map[string]interface{}{
					"prev-page":    (*int)(nil),
					"current-page": 1,
					"next-page":    (*int)(nil),
					"total-count":  5,
					"total-pages":  1,
				},
			},
		},
		{
			name:      "multiple pages",
			paginated: NewPaginatedMock("/v2/api/foobar", 101, 3, 20),
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
			pagination := NewPagination(tt.paginated)
			assert.Equal(t, tt.wantLinks, pagination.JSONAPIPaginationLinks())
			assert.Equal(t, tt.wantMeta, pagination.JSONAPIPaginationMeta())
		})
	}
}
