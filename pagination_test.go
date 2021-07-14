package ots

import (
	"testing"

	tfe "github.com/leg100/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name  string
		opts  tfe.ListOptions
		count int
		want  tfe.Pagination
	}{
		{
			name:  "one page",
			opts:  tfe.ListOptions{PageNumber: 1, PageSize: 20},
			count: 5,
			want: tfe.Pagination{
				CurrentPage:  1,
				PreviousPage: 1,
				NextPage:     1,
				TotalCount:   5,
				TotalPages:   1,
			},
		},
		{
			name:  "multiple pages",
			opts:  tfe.ListOptions{PageNumber: 3, PageSize: 20},
			count: 101,
			want: tfe.Pagination{
				CurrentPage:  3,
				PreviousPage: 2,
				NextPage:     4,
				TotalCount:   101,
				TotalPages:   6,
			},
		},
		{
			name:  "no results",
			opts:  tfe.ListOptions{PageNumber: 1, PageSize: 20},
			count: 0,
			want: tfe.Pagination{
				CurrentPage:  1,
				PreviousPage: 1,
				NextPage:     1,
				TotalCount:   0,
				TotalPages:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagination := NewPagination(tt.opts, tt.count)
			assert.Equal(t, tt.want, *pagination)
		})
	}
}
