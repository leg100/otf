package ots

import (
	"math"

	tfe "github.com/leg100/go-tfe"
)

// NewPagination constructs a Pagination obj.
func NewPagination(opts tfe.ListOptions, count int) *tfe.Pagination {
	return &tfe.Pagination{
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

func previousPage(currentPage int) int {
	if currentPage > 1 {
		return currentPage - 1
	}
	return 1
}

func nextPage(opts tfe.ListOptions, count int) int {
	if opts.PageNumber < totalPages(count, opts.PageSize) {
		return opts.PageNumber + 1
	}
	return 1
}
