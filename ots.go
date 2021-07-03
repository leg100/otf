package ots

import (
	"math/rand"
	"time"

	"github.com/gorilla/schema"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 100

	DefaultUserID   = "user-123"
	DefaultUsername = "ots"

	alphanumeric = "abcdefghijkmnopqrstuvwxyzABCDEFGHIJKMNOPQRSTUVWXYZ0123456789"
)

var encoder = schema.NewEncoder()

func String(str string) *string { return &str }
func Int(i int) *int            { return &i }

func GenerateRandomString(size int) string {
	// Without this, Go would generate the same random sequence each run.
	rand.Seed(time.Now().UnixNano())

	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return string(buf)
}

// ListOptions is used to specify pagination options when making API requests.
// Pagination allows breaking up large result sets into chunks, or "pages".
type ListOptions struct {
	// The page number to request. The results vary based on the PageSize.
	PageNumber int `schema:"page[number]"`

	// The number of elements returned in a single page.
	PageSize int `schema:"page[size]"`
}

// GetPageNumber partially implements the Paginated interface
func (o *ListOptions) GetPageNumber() int {
	return o.PageNumber
}

// GetPageSize partially implements the Paginated interface
func (o *ListOptions) GetPageSize() int {
	return o.PageSize
}

// Sanitize list options' values, setting defaults and ensuring they adhere to
// mins/maxs.
func (o *ListOptions) Sanitize() {
	if o.PageNumber <= 0 {
		o.PageNumber = DefaultPageNumber
	}

	if o.PageSize <= 0 {
		o.PageSize = DefaultPageSize
	} else if o.PageSize > 100 {
		o.PageSize = MaxPageSize
	}
}
