package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPagination constructs a pagination object from list options and a count
// and then runs a series of tests on the object.
func TestPagination(t *testing.T) {
	tests := []struct {
		name             string
		opts             ListOptions
		count            int
		wantCurrentPage  int
		wantPreviousPage *int
		wantNextPage     *int
		wantTotalCount   int
		wantTotalPages   int
	}{
		{
			name:             "one page",
			opts:             ListOptions{PageNumber: 1, PageSize: 20},
			count:            5,
			wantCurrentPage:  1,
			wantPreviousPage: nil,
			wantNextPage:     nil,
			wantTotalCount:   5,
			wantTotalPages:   1,
		},
		{
			name:             "multiple pages",
			opts:             ListOptions{PageNumber: 3, PageSize: 20},
			count:            101,
			wantCurrentPage:  3,
			wantPreviousPage: Int(2),
			wantNextPage:     Int(4),
			wantTotalCount:   101,
			wantTotalPages:   6,
		},
		{
			name:             "no results",
			opts:             ListOptions{PageNumber: 1, PageSize: 20},
			count:            0,
			wantCurrentPage:  1,
			wantPreviousPage: nil,
			wantNextPage:     nil,
			wantTotalCount:   0,
			wantTotalPages:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagination := NewPagination(tt.opts, tt.count)
			assert.Equal(t, tt.wantCurrentPage, pagination.CurrentPage())
			assert.Equal(t, tt.wantPreviousPage, pagination.PrevPage())
			assert.Equal(t, tt.wantNextPage, pagination.NextPage())
			assert.Equal(t, tt.wantTotalCount, pagination.TotalCount())
			assert.Equal(t, tt.wantTotalPages, pagination.TotalPages())
		})
	}
}

func TestListOptions(t *testing.T) {
	opts := ListOptions{}
	assert.Equal(t, 0, opts.GetOffset())
	assert.Equal(t, 100, opts.GetLimit())

	opts = ListOptions{PageNumber: 1, PageSize: 20}
	assert.Equal(t, 0, opts.GetOffset())
	assert.Equal(t, 20, opts.GetLimit())

	opts = ListOptions{PageNumber: 5, PageSize: 20}
	assert.Equal(t, 80, opts.GetOffset())
	assert.Equal(t, 20, opts.GetLimit())
}
