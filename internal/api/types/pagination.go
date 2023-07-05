package types

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type (
	// ListOptions is used to specify pagination options when making API requests.
	// Pagination allows breaking up large result sets into chunks, or "pages".
	ListOptions struct {
		// The page number to request. The results vary based on the PageSize.
		PageNumber int `schema:"page[number],omitempty"`
		// The number of elements returned in a single page.
		PageSize int `schema:"page[size],omitempty"`
	}

	// Pagination is used to return the pagination details of an API request.
	Pagination struct {
		CurrentPage  int  `json:"current-page"`
		PreviousPage *int `json:"prev-page"`
		NextPage     *int `json:"next-page"`
		TotalPages   int  `json:"total-pages"`
		TotalCount   int  `json:"total-count"`
	}
)
