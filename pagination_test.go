package ots

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name  string
		opts  ListOptions
		count int
		want  Pagination
	}{
		{
			name:  "one page",
			opts:  ListOptions{PageNumber: 1, PageSize: 20},
			count: 5,
			want: Pagination{
				CurrentPage:  1,
				PreviousPage: (*int)(nil),
				NextPage:     (*int)(nil),
				TotalCount:   5,
				TotalPages:   1,
			},
		},
		{
			name:  "multiple pages",
			opts:  ListOptions{PageNumber: 3, PageSize: 20},
			count: 101,
			want: Pagination{
				CurrentPage:  3,
				PreviousPage: Int(2),
				NextPage:     Int(4),
				TotalCount:   101,
				TotalPages:   6,
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
