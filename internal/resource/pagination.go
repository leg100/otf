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

	Page[T any] struct {
		Items []T
		*Pagination
	}
)

func NewPage[T any](resources []T, opts PageOptions, count int64) *Page[T] {
	return &Page[T]{
		Items:      resources,
		Pagination: NewPagination(opts, count),
	}
}

// Paginate paginates a slice of resources, returning a slice for the current
// page and a pagination meta object. The slice should be the entire result set.
func Paginate[S comparable](resources []S, opts PageOptions) *Page[S] {
	opts = opts.normalize()
	count := int64(len(resources))

	// remove items outside the current page
	if end := opts.PageSize * opts.PageNumber; len(resources) > end {
		resources = resources[:end]
	}
	start := opts.PageSize * (opts.PageNumber - 1)
	if start > len(resources) {
		// paging is out-of-range: return empty list
		return &Page[S]{}
	}
	resources = resources[start:]

	return NewPage[S](resources, opts, count)
}

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
