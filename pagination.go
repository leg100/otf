package otf

import (
	"math"
	"net/url"
	"strconv"

	jsonapi "github.com/leg100/otf/http/dto"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 10
	MaxPageSize       = 100
)

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	// list options from API request
	opts ListOptions
	// total unpaginated count
	count int
}

func (p *Pagination) CurrentPage() int { return p.opts.PageNumber }
func (p *Pagination) TotalCount() int  { return p.count }

// PrevPage returns the previous page number or nil if there isn't one.
func (p *Pagination) PrevPage() *int {
	if p.opts.sanitizedPageNumber() > 1 {
		return Int(p.opts.prevPage())
	}
	return nil
}

// NextPage returns the next page number or nil if there isn't one.
func (p *Pagination) NextPage() *int {
	if p.opts.sanitizedPageNumber() < p.TotalPages() {
		return Int(p.opts.nextPage())
	}
	return nil
}

func (p *Pagination) TotalPages() int {
	pages := float64(p.count) / float64(p.opts.sanitizedPageSize())
	// total must be a round number greater than 0
	return int(math.Max(1, math.Ceil(pages)))
}

// ToJSONAPI assembles a JSON-API DTO for wire serialization.
func (p *Pagination) ToJSONAPI() *jsonapi.Pagination {
	return &jsonapi.Pagination{
		CurrentPage:  p.opts.sanitizedPageNumber(),
		PreviousPage: p.PrevPage(),
		NextPage:     p.NextPage(),
		TotalPages:   p.TotalPages(),
		TotalCount:   p.count,
	}
}

// NewPagination constructs a Pagination obj.
func NewPagination(opts ListOptions, count int) *Pagination {
	return &Pagination{opts, count}
}

// UnmarshalPaginationJSONAPI converts a JSON-API DTO into a pagination object.
func UnmarshalPaginationJSONAPI(json *jsonapi.Pagination) *Pagination {
	return &Pagination{
		count: json.TotalCount,
		// we can't determine the page size so we'll just pass in 0 which
		// NewListOptions interprets as the default page size
		opts: ListOptions{PageNumber: json.CurrentPage, PageSize: 0},
	}
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
	return (o.sanitizedPageNumber() - 1) * o.sanitizedPageSize()
}

// GetLimit calculates the limit for use in SQL queries.
func (o ListOptions) GetLimit() int {
	return o.sanitizedPageSize()
}

// NextPageQuery produces query params for the next page
func (o ListOptions) NextPageQuery() string {
	q := url.Values{}
	q.Add("page[number]", strconv.Itoa(o.nextPage()))
	q.Add("page[size]", strconv.Itoa(o.sanitizedPageSize()))
	return q.Encode()
}

// PrevPageQuery produces query params for the previous page
func (o ListOptions) PrevPageQuery() string {
	q := url.Values{}
	q.Add("page[number]", strconv.Itoa(o.prevPage()))
	q.Add("page[size]", strconv.Itoa(o.sanitizedPageSize()))
	return q.Encode()
}

func (o ListOptions) nextPage() int {
	return o.sanitizedPageNumber() + 1
}

func (o ListOptions) prevPage() int {
	return o.sanitizedPageNumber() - 1
}

// sanitizedPageNumber is the page number following sanitization.
func (o ListOptions) sanitizedPageNumber() int {
	if o.PageNumber == 0 {
		return 1
	}
	return o.PageNumber
}

// sanitizedPageSize is the page size following sanitization.
func (o ListOptions) sanitizedPageSize() int {
	if o.PageSize == 0 {
		return DefaultPageSize
	}
	if o.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return o.PageSize
}
