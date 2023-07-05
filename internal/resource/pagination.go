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
	// Page is a segment of a result set.
	Page[T any] struct {
		Items []T
		*Pagination
	}

	// Pagination provides metadata about a page.
	Pagination struct {
		CurrentPage  int  `json:"current-page"`
		PreviousPage *int `json:"prev-page"`
		NextPage     *int `json:"next-page"`
		TotalPages   int  `json:"total-pages"`
		TotalCount   int  `json:"total-count"`
	}

	// PageOptions are used to request a specific page.
	PageOptions struct {
		// The page number to request. The results vary based on the PageSize.
		PageNumber int `schema:"page[number],omitempty"`
		// The number of elements returned in a single page.
		PageSize int `schema:"page[size],omitempty"`
	}
)

// NewPage constructs a page from a list of resources. If the list argument
// represents the full result set then count should be nil; if count is non-nil
// then the list is deemed to be a segment of a result set and count is the size
// of the full result set. This latter case is useful, say, if a database has
// already produced a segment of a full result set, e.g. using LIMIT and OFFSET.
func NewPage[T any](resources []T, opts PageOptions, count *int64) *Page[T] {
	opts = opts.normalize()

	var metadata *Pagination

	if count == nil {
		metadata = newPagination(opts, int64(len(resources)))
		// remove items outside the current page
		if end := opts.PageSize * opts.PageNumber; len(resources) > end {
			resources = resources[:end]
		}
		if start := opts.PageSize * (opts.PageNumber - 1); start > len(resources) {
			// paging is out-of-range: return empty list
			resources = []T{}
		} else {
			resources = resources[start:]
		}
	} else {
		metadata = newPagination(opts, *count)
	}

	return &Page[T]{
		Items:      resources,
		Pagination: metadata,
	}
}

// ListAll is a helper for retrieving all pages. The provided fn should perform
// an operation that retrieves a page at a time.
func ListAll[T any](fn func(PageOptions) (*Page[T], error)) ([]T, error) {
	var (
		opts PageOptions
		all  []T
	)
	for {
		page, err := fn(opts)
		if err != nil {
			return nil, err
		}
		// should never happen...
		if page == nil || page.Items == nil {
			break
		}
		all = append(all, page.Items...)
		if page.NextPage == nil {
			break
		}
		opts.PageNumber = *page.NextPage
	}
	return all, nil
}

func newPagination(opts PageOptions, count int64) *Pagination {
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
