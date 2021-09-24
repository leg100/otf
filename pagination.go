package otf

import (
	"math"
)

// SanitizeListOptions ensures list options adhere to mins and maxs
func SanitizeListOptions(opts *ListOptions) {
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

// NewPagination constructs a Pagination obj.
func NewPagination(opts ListOptions, count int) *Pagination {
	SanitizeListOptions(&opts)

	return &Pagination{
		CurrentPage:  opts.PageNumber,
		PreviousPage: previousPage(opts.PageNumber),
		NextPage:     nextPage(opts, count),
		TotalPages:   totalPages(count, opts.PageSize),
		TotalCount:   count,
	}
}

func totalPages(totalCount, pageSize int) int {
	return int(math.Max(1, math.Ceil(float64(totalCount)/float64(pageSize))))
}

func previousPage(currentPage int) int {
	if currentPage > 1 {
		return currentPage - 1
	}
	return 1
}

func nextPage(opts ListOptions, count int) int {
	if opts.PageNumber < totalPages(count, opts.PageSize) {
		return opts.PageNumber + 1
	}
	return 1
}
