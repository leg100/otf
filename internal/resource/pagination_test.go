package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				PreviousPage: new(2),
				NextPage:     new(4),
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
			got := newPagination(tt.opts, tt.count)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewPage(t *testing.T) {
	// construct a slice of numbers from 1 through 101
	s := make([]int, 101)
	for i := range s {
		s[i] = i + 1
	}

	tests := []struct {
		name  string
		opts  PageOptions
		count *int64
		want  Page[int]
	}{
		{
			"default",
			PageOptions{},
			nil,
			Page[int]{
				Items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				Pagination: &Pagination{
					CurrentPage: 1,
					TotalCount:  101,
					TotalPages:  6,
					NextPage:    new(2),
				},
			},
		},
		{
			"second page",
			PageOptions{PageSize: 10, PageNumber: 2},
			nil,
			Page[int]{
				Items: []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				Pagination: &Pagination{
					CurrentPage:  2,
					TotalCount:   101,
					TotalPages:   11,
					PreviousPage: new(1),
					NextPage:     new(3),
				},
			},
		},
		{
			"last page",
			PageOptions{PageSize: 10, PageNumber: 11},
			nil,
			Page[int]{
				Items: []int{101},
				Pagination: &Pagination{
					CurrentPage:  11,
					TotalCount:   101,
					TotalPages:   11,
					PreviousPage: new(10),
				},
			},
		},
		{
			"out of range",
			PageOptions{PageSize: 10, PageNumber: 99},
			nil,
			Page[int]{
				Items: []int{},
				Pagination: &Pagination{
					CurrentPage:  99,
					TotalCount:   101,
					TotalPages:   11,
					PreviousPage: new(98),
				},
			},
		},
		{
			"page from database",
			PageOptions{PageSize: 100, PageNumber: 1},
			new(int64(201)),
			Page[int]{
				// note s is now a segment within a larger result set of 201
				// items.
				Items: s,
				Pagination: &Pagination{
					CurrentPage: 1,
					TotalCount:  201,
					TotalPages:  3,
					NextPage:    new(2),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPage(s, tt.opts, tt.count)
			assert.Equal(t, &tt.want, got)
		})
	}
}

func TestListAll(t *testing.T) {
	type foo int

	var page int
	got, err := ListAll(func(opts PageOptions) (*Page[foo], error) {
		if opts.PageNumber == 10 {
			return &Page[foo]{
				Pagination: &Pagination{},
			}, nil
		}
		page++
		return &Page[foo]{
			Items: []foo{foo(page)},
			Pagination: &Pagination{
				NextPage: new(page),
			},
		}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, []foo{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, got)
}

func TestListAll_infinite_pagination(t *testing.T) {
	type foo int

	// This demonstrates an incorrectly implemented func being passed to
	// ListAll, which retrieves the same page ad infinitum.
	_, err := ListAll(func(opts PageOptions) (*Page[foo], error) {
		return &Page[foo]{
			Items: []foo{0},
			Pagination: &Pagination{
				NextPage: new(1),
			},
		}, nil
	})
	assert.Equal(t, ErrInfinitePaginationDetected, err)
}
