package internal

import (
	"math"
	"net/url"
	"strconv"

	"github.com/leg100/otf/internal/api/types"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = MaxPageSize
	MaxPageSize       = 100
)

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	// list options from API request
	Opts ListOptions
	// total unpaginated count
	Count int
}

// NewPaginationFromJSONAPI constructs pagination from a json:api struct
func NewPaginationFromJSONAPI(json *types.Pagination) *Pagination {
	return &Pagination{
		Count: json.TotalCount,
		// we can't determine the page size so we'll just pass in 0 which
		// ListOptions interprets as the default page size
		Opts: ListOptions{PageNumber: json.CurrentPage, PageSize: 0},
	}
}

func (p *Pagination) CurrentPage() int { return p.Opts.PageNumber }
func (p *Pagination) TotalCount() int  { return p.Count }

// PrevPage returns the previous page number or nil if there isn't one.
func (p *Pagination) PrevPage() *int {
	if p.Opts.SanitizedPageNumber() > 1 {
		return Int(p.Opts.prevPage())
	}
	return nil
}

// NextPage returns the next page number or nil if there isn't one.
func (p *Pagination) NextPage() *int {
	if p.Opts.SanitizedPageNumber() < p.TotalPages() {
		return Int(p.Opts.nextPage())
	}
	return nil
}

func (p *Pagination) TotalPages() int {
	pages := float64(p.Count) / float64(p.Opts.SanitizedPageSize())
	// total must be a round number greater than 0
	return int(math.Max(1, math.Ceil(pages)))
}

// NextPageQuery produces query params for the next page
func (p *Pagination) NextPageQuery() string {
	q := url.Values{}
	q.Add("page[number]", strconv.Itoa(p.Opts.nextPage()))
	q.Add("page[size]", strconv.Itoa(p.Opts.SanitizedPageSize()))
	return q.Encode()
}

// PrevPageQuery produces query params for the previous page
func (p *Pagination) PrevPageQuery() string {
	q := url.Values{}
	q.Add("page[number]", strconv.Itoa(p.Opts.prevPage()))
	q.Add("page[size]", strconv.Itoa(p.Opts.SanitizedPageSize()))
	return q.Encode()
}

// NewPagination constructs a Pagination obj.
func NewPagination(opts ListOptions, count int) *Pagination {
	return &Pagination{opts, count}
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
func (o ListOptions) GetOffset() int {
	return (o.SanitizedPageNumber() - 1) * o.SanitizedPageSize()
}

// GetLimit calculates the limit for use in SQL queries.
func (o ListOptions) GetLimit() int {
	return o.SanitizedPageSize()
}

func (o ListOptions) nextPage() int {
	return o.SanitizedPageNumber() + 1
}

func (o ListOptions) prevPage() int {
	return o.SanitizedPageNumber() - 1
}

// SanitizedPageNumber is the page number following sanitization.
func (o ListOptions) SanitizedPageNumber() int {
	if o.PageNumber == 0 {
		return 1
	}
	return o.PageNumber
}

// SanitizedPageSize is the page size following sanitization.
func (o ListOptions) SanitizedPageSize() int {
	if o.PageSize == 0 {
		return DefaultPageSize
	}
	if o.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return o.PageSize
}
