package resource

import (
	"errors"
	"math"
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
		PageNumber int `schema:"page,omitempty"`
		// The number of elements returned in a single page.
		PageSize int `schema:"page_size,omitempty"`
	}
)

// NewPage constructs a page from a list of resources. If the list argument
// represents the full result set then count should be nil; if count is non-nil
// then the list is deemed to be a segment of a result set and count is the size
// of the full result set. This latter case is useful, say, if a database has
// already produced a segment of a full result set, e.g. using LIMIT and OFFSET.
func NewPage[T any](resources []T, opts PageOptions, count *int64) *Page[T] {
	opts = opts.Normalize()

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

var ErrInfinitePaginationDetected = errors.New("infinite pagination detected")

// ListAll is a helper for retrieving all pages. The provided fn should perform
// an operation that retrieves a page at a time.
func ListAll[T any](fn func(PageOptions) (*Page[T], error)) ([]T, error) {
	var (
		opts = PageOptions{PageSize: MaxPageSize}
		all  []T
		// keep track of the last NextPage to prevent an infinite loop.
		lastNextPage int
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
		if *page.NextPage == lastNextPage {
			return nil, ErrInfinitePaginationDetected
		}
		opts.PageNumber = *page.NextPage
		lastNextPage = *page.NextPage
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
		pagination.PreviousPage = new(opts.PageNumber - 1)
	}
	if opts.PageNumber < pagination.TotalPages {
		pagination.NextPage = new(opts.PageNumber + 1)
	}

	return &pagination
}

// Normalize page number and size
func (o PageOptions) Normalize() PageOptions {
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
