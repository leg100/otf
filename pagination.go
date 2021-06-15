package ots

import (
	"math"
	"net/url"
	"reflect"

	"github.com/google/jsonapi"
)

type Paginated interface {
	GetItems() interface{}
	GetListOptions() ListOptions
	GetPath() string
}

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	// Current page
	current int

	// Maximum number of items on each page
	pageSize int

	// Total number of items
	totalItems int

	// Path to use in JSON-API links
	path string
}

// NewPagination constructs a Pagination obj.
func NewPagination(p Paginated) *Pagination {
	items := p.GetItems()
	totalItems := len(convertToSliceInterface(&items))

	return &Pagination{
		current:    p.GetListOptions().PageNumber,
		pageSize:   p.GetListOptions().PageSize,
		totalItems: totalItems,
		path:       p.GetPath(),
	}
}

func (p *Pagination) JSONAPIPaginationLinks() *jsonapi.Links {
	linksmap := map[string]interface{}{
		"self":  p.link(p.current),
		"first": p.link(1),
		"last":  p.link(p.totalPages()),
	}

	if prev := p.prev(); prev != nil {
		linksmap["prev"] = p.link(*prev)
	}

	if next := p.next(); next != nil {
		linksmap["next"] = p.link(*next)
	}

	links := jsonapi.Links(linksmap)
	return &links
}

func (p *Pagination) JSONAPIPaginationMeta() *jsonapi.Meta {
	m := map[string]interface{}{
		"current-page": p.current,
		"total-pages":  p.totalPages(),
		"total-count":  p.totalItems,
		"prev-page":    p.prev(),
		"next-page":    p.next(),
	}
	return &jsonapi.Meta{"pagination": m}
}

func (p *Pagination) link(number int) string {
	opts := ListOptions{PageNumber: number, PageSize: p.pageSize}
	query := url.Values{}
	if err := encoder.Encode(opts, query); err != nil {
		panic(err.Error())
	}

	return (&url.URL{Path: p.path, RawQuery: query.Encode()}).String()
}

func (p *Pagination) totalPages() int {
	return int(math.Ceil(float64(p.totalItems) / float64(p.pageSize)))
}

func (p *Pagination) prev() *int {
	if p.current > 1 {
		return Int(p.current - 1)
	}
	return nil
}

func (p *Pagination) next() *int {
	if p.current < p.totalPages() {
		return Int(p.current + 1)
	}
	return nil
}

func convertToSliceInterface(i *interface{}) []interface{} {
	vals := reflect.ValueOf(*i)
	if vals.Kind() != reflect.Slice {
		panic(jsonapi.ErrExpectedSlice.Error())
	}
	var response []interface{}
	for x := 0; x < vals.Len(); x++ {
		response = append(response, vals.Index(x).Interface())
	}
	return response
}
