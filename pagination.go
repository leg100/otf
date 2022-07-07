package otf

import (
	"math"
	"net/url"
	"strconv"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 10
	MaxPageSize       = 100
)

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	CurrentPage  int
	PreviousPage *int
	NextPage     *int
	TotalPages   int
	TotalCount   int
}

func (p *Pagination) HasNextPage() bool { return p.NextPage != nil }
func (p *Pagination) HasPrevPage() bool { return p.PreviousPage != nil }

func (p *Pagination) NextPageQuery() string {
	if p.NextPage == nil {
		return ""
	}
	q := url.Values{}
	q.Add("page[number]", strconv.Itoa(*p.NextPage))
	q.Add("page[size]", strconv.Itoa(DefaultPageSize))
	return q.Encode()
}

func (p *Pagination) PrevPageQuery() string {
	if p.PreviousPage == nil {
		return ""
	}
	q := url.Values{}
	q.Add("page[number]", strconv.Itoa(*p.PreviousPage))
	q.Add("page[size]", strconv.Itoa(DefaultPageSize))
	return q.Encode()
}

// ListOptions is used to specify pagination options when making API requests.
// Pagination allows breaking up large result sets into chunks, or "pages".
type ListOptions struct {
	// The page number to request. The results vary based on the PageSize.
	PageNumber int `schema:"page[number],omitempty"`

	// The number of elements returned in a single page.
	PageSize int `schema:"page[size],omitempty"`
}

// GetOffset calculates the offset for use in SQL queries.
func (o *ListOptions) GetOffset() int {
	if o.PageNumber == 0 {
		return 0
	}

	return (o.PageNumber - 1) * o.PageSize
}

// GetLimit calculates the limit for use in SQL queries.
func (o *ListOptions) GetLimit() int {
	// TODO: remove MaxPageSize - this is too complicated
	if o.PageSize == 0 {
		return math.MaxInt
	} else if o.PageSize > MaxPageSize {
		return MaxPageSize
	}

	return o.PageSize
}

// NewPagination constructs a Pagination obj.
func NewPagination(opts ListOptions, count int) *Pagination {
	sanitizeListOptions(&opts)

	return &Pagination{
		CurrentPage:  opts.PageNumber,
		PreviousPage: previousPage(opts.PageNumber),
		NextPage:     nextPage(opts, count),
		TotalPages:   totalPages(count, opts.PageSize),
		TotalCount:   count,
	}
}

// sanitizeListOptions ensures list options adhere to mins and maxs
func sanitizeListOptions(opts *ListOptions) {
	if opts.PageNumber == 0 {
		opts.PageNumber = 1
	}

	switch {
	case opts.PageSize > 100:
		opts.PageSize = MaxPageSize
	case opts.PageSize <= 0:
		opts.PageSize = DefaultPageSize
	}
}

func totalPages(totalCount, pageSize int) int {
	return int(math.Max(1, math.Ceil(float64(totalCount)/float64(pageSize))))
}

func previousPage(currentPage int) *int {
	if currentPage > 1 {
		return Int(currentPage - 1)
	}
	return nil
}

func nextPage(opts ListOptions, count int) *int {
	if opts.PageNumber < totalPages(count, opts.PageSize) {
		return Int(opts.PageNumber + 1)
	}
	return nil
}
