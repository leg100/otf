package ots

import (
	"math"
)

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	CurrentPage  int  `json:"current-page"`
	PreviousPage *int `json:"prev-page"`
	NextPage     *int `json:"next-page"`
	TotalPages   int  `json:"total-pages"`
	TotalCount   int  `json:"total-count"`
}

// NewPagination constructs a Pagination obj.
func NewPagination(opts ListOptions, count int) *Pagination {
	return &Pagination{
		CurrentPage:  opts.PageNumber,
		PreviousPage: previousPage(opts.PageNumber),
		NextPage:     nextPage(opts, count),
		TotalPages:   totalPages(count, opts.PageSize),
		TotalCount:   count,
	}
}

func totalPages(totalCount, pageSize int) int {
	return int(math.Ceil(float64(totalCount) / float64(pageSize)))
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
