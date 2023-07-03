package resource

import (
	"math"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 100
)

type (
	// Pagination is used to return the pagination details of an API request.
	Pagination struct {
		CurrentPage  int  `json:"current-page"`
		PreviousPage *int `json:"prev-page"`
		NextPage     *int `json:"next-page"`
		TotalPages   int  `json:"total-pages"`
		TotalCount   int  `json:"total-count"`
	}

	// PageOptions are used to specify pagination options when making API requests.
	// Pagination allows breaking up large result sets into chunks, or "pages".
	PageOptions struct {
		// The page number to request. The results vary based on the PageSize.
		PageNumber int `schema:"page[number],omitempty"`
		// The number of elements returned in a single page.
		PageSize int `schema:"page[size],omitempty"`
	}
)

func NewPagination(opts PageOptions, count int64) *Pagination {
	opts = opts.normalize()

	pagination := Pagination{
		CurrentPage: opts.PageNumber,
		TotalCount:  int(count),
	}

	// total pages must be a round number greater than 0
	pages := float64(count) / float64(opts.PageSize)
	pagination.TotalPages = int(math.Max(1, math.Ceil(pages)))

	if opts.PageNumber > 1 {
		pagination.PreviousPage = internal.Int(opts.PageNumber - 1)
	}
	if opts.PageNumber < pagination.TotalPages {
		pagination.NextPage = internal.Int(opts.PageNumber + 1)
	}

	return &pagination
}

// Paginate paginates a slice of resources, returning a slice for the current
// page and a pagination meta object. The slice should be the entire result set,
// and r should contain pagination query parameters.
func Paginate[S comparable](from []S, opts PageOptions) ([]S, *Pagination) {
	pagination := NewPagination(opts, int64(len(from)))
	opts = opts.normalize()

	// remove items outside the current page
	end := opts.PageSize * opts.PageNumber
	if len(from) > end {
		from = from[:end]
	}

	start := opts.PageSize * (opts.PageNumber - 1)
	if start > len(from) {
		// paging is out-of-range: return empty list
		return []S{}, pagination
	}
	from = from[start:]

	return from, pagination
}

// GetOffset calculates the offset for use in SQL queries.
func (o PageOptions) GetOffset() pgtype.Int8 {
	o = o.normalize()
	return sql.Int8((o.PageNumber - 1) * o.PageSize)
}

// GetLimit calculates the limit for use in SQL queries.
func (o PageOptions) GetLimit() pgtype.Int8 {
	return sql.Int8(o.normalize().PageSize)
}

// normalize page number and size
func (o PageOptions) normalize() PageOptions {
	if o.PageNumber < 1 {
		o.PageNumber = 1
	}
	if o.PageSize <= 0 {
		o.PageSize = DefaultPageSize
	}
	if o.PageSize > MaxPageSize {
		o.PageSize = MaxPageSize
	}
	return o
}
