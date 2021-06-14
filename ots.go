package ots

import (
	"math"
	"net/url"
	"strconv"

	"github.com/google/jsonapi"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 100
)

type Paginated interface {
	GetItems() interface{}
	JSONAPIPaginationLinks() *jsonapi.Links
	JSONAPIPaginationMeta() *jsonapi.Meta
}

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	CurrentPage  int  `json:"current-page"`
	PreviousPage *int `json:"prev-page,omitempty"`
	NextPage     *int `json:"next-page,omitempty"`
	TotalPages   int  `json:"total-pages"`
	TotalCount   int  `json:"total-count"`

	PageSize int `json:"-"`

	path string
}

// ListOptions is used to specify pagination options when making API requests.
// Pagination allows breaking up large result sets into chunks, or "pages".
type ListOptions struct {
	// The page number to request. The results vary based on the PageSize.
	PageNumber int `schema:"page[number]"`

	// The number of elements returned in a single page.
	PageSize int `schema:"page[size]"`
}

func (o *ListOptions) Sanitize() {
	if o.PageNumber == 0 {
		o.PageNumber = DefaultPageNumber
	}

	if o.PageSize == 0 {
		o.PageSize = DefaultPageSize
	} else if o.PageSize > 100 {
		o.PageSize = MaxPageSize
	}
}

func NewPagination(path string, opts ListOptions, count int) *Pagination {
	pagination := &Pagination{
		CurrentPage: opts.PageNumber,
		TotalPages:  int(math.Ceil(float64(count) / float64(opts.PageSize))),
		TotalCount:  count,
		PageSize:    opts.PageSize,
		path:        path,
	}

	if pagination.CurrentPage < pagination.TotalPages {
		pagination.NextPage = Int(pagination.CurrentPage + 1)
	}
	if pagination.CurrentPage > 1 {
		pagination.PreviousPage = Int(pagination.CurrentPage - 1)
	}
	return pagination
}

func (p *Pagination) JSONAPIPaginationLinks() *jsonapi.Links {
	linksmap := map[string]interface{}{
		"self":  p.link(p.CurrentPage),
		"first": p.link(1),
		"last":  p.link(p.TotalPages),
	}

	if p.PreviousPage != nil {
		linksmap["prev"] = p.link(*p.PreviousPage)
	}

	if p.NextPage != nil {
		linksmap["next"] = p.link(*p.NextPage)
	}

	links := jsonapi.Links(linksmap)
	return &links
}

func (p *Pagination) JSONAPIPaginationMeta() *jsonapi.Meta {
	metamap := map[string]interface{}{
		"current-page": p.CurrentPage,
		"total-pages":  p.TotalPages,
		"total-count":  p.TotalCount,
	}

	if p.PreviousPage != nil {
		metamap["prev-page"] = p.PreviousPage
	}
	if p.NextPage != nil {
		metamap["next-page"] = p.NextPage
	}

	return &jsonapi.Meta{
		"pagination": metamap,
	}
}

func ListOptionsFromQuery(query url.Values) (*ListOptions, error) {
	opts := &ListOptions{
		PageNumber: 1,
		PageSize:   DefaultPageSize,
	}

	if num := query.Get("page[number]"); num != "" {
		num, err := strconv.ParseInt(num, 10, 0)
		if err != nil {
			return nil, err
		}
		opts.PageNumber = int(num)
	}

	if size := query.Get("page[size]"); size != "" {
		size, err := strconv.ParseInt(size, 10, 0)
		if err != nil {
			return nil, err
		}
		opts.PageSize = int(size)
	}

	return opts, nil
}

func Int(i int) *int { return &i }

func (p *Pagination) link(number int) string {
	query := &url.Values{}
	query.Set("page[number]", strconv.Itoa(number))
	query.Set("page[size]", strconv.Itoa(p.PageSize))

	return (&url.URL{Path: p.path, RawQuery: query.Encode()}).String()
}
