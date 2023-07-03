package resource

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name  string
		opts  PageOptions
		count int64
		want  *Pagination
	}{
		{
			name:  "one page",
			opts:  PageOptions{PageNumber: 1, PageSize: 20},
			count: 5,
			want: &Pagination{
				CurrentPage:  1,
				PreviousPage: nil,
				NextPage:     nil,
				TotalCount:   5,
				TotalPages:   1,
			},
		},
		{
			name:  "multiple pages",
			opts:  PageOptions{PageNumber: 3, PageSize: 20},
			count: 101,
			want: &Pagination{
				CurrentPage:  3,
				PreviousPage: internal.Int(2),
				NextPage:     internal.Int(4),
				TotalCount:   101,
				TotalPages:   6,
			},
		},
		{
			name:  "no results",
			opts:  PageOptions{PageNumber: 1, PageSize: 20},
			count: 0,
			want: &Pagination{
				CurrentPage:  1,
				PreviousPage: nil,
				NextPage:     nil,
				TotalCount:   0,
				TotalPages:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPagination(tt.opts, tt.count)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaginate(t *testing.T) {
	// construct a slice of numbers from 1 through 101
	s := make([]int, 101)
	for i := 0; i < len(s); i++ {
		s[i] = i + 1
	}
	tests := []struct {
		name           string
		opts           PageOptions
		want           []int
		wantPagination Pagination
	}{
		{
			"default",
			PageOptions{},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			Pagination{
				CurrentPage: 1,
				TotalCount:  101,
				TotalPages:  6,
				NextPage:    internal.Int(2),
			},
		},
		{
			"second page",
			PageOptions{PageSize: 10, PageNumber: 2},
			[]int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			Pagination{
				CurrentPage:  2,
				TotalCount:   101,
				TotalPages:   11,
				PreviousPage: internal.Int(1),
				NextPage:     internal.Int(3),
			},
		},
		{
			"last page",
			PageOptions{PageSize: 10, PageNumber: 11},
			[]int{101},
			Pagination{
				CurrentPage:  11,
				TotalCount:   101,
				TotalPages:   11,
				PreviousPage: internal.Int(10),
			},
		},
		{
			"out of range",
			PageOptions{PageSize: 10, PageNumber: 99},
			[]int{},
			Pagination{
				CurrentPage:  99,
				TotalCount:   101,
				TotalPages:   11,
				PreviousPage: internal.Int(98),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, opt := Paginate(s, tt.opts)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, &tt.wantPagination, opt)
		})
	}
}
