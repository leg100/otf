/*
Package otf is responsible for domain logic.
*/
package otf

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 100

	DefaultUserID   = "user-123"
	DefaultUsername = "otf"

	alphanumeric = "abcdefghijkmnopqrstuvwxyzABCDEFGHIJKMNOPQRSTUVWXYZ0123456789"
)

// A regular expression used to validate common string ID patterns.
var reStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)

// A regular expression used to validate semantic versions (major.minor.patch).
var reSemanticVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

func String(str string) *string { return &str }
func Int(i int) *int            { return &i }
func Int64(i int64) *int64      { return &i }
func UInt(i uint) *uint         { return &i }

// TimeNow is a convenience func to return the pointer of the current time
func TimeNow() *time.Time {
	t := time.Now()
	return &t
}

// GenerateID generates a unique identifier with the given prefix
func GenerateID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, GenerateRandomString(16))
}

// GenerateRandomString generates a random string composed of alphanumeric
// characters of length size.
func GenerateRandomString(size int) string {
	// Without this, Go would generate the same random sequence each run.
	rand.Seed(time.Now().UnixNano())

	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return string(buf)
}

// Resources summaries updates to a workspace's resources, either proposed as
// part of a plan, or made as a result of an apply.
type Resources struct {
	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
}

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	CurrentPage  int `json:"current-page"`
	PreviousPage int `json:"prev-page"`
	NextPage     int `json:"next-page"`
	TotalPages   int `json:"total-pages"`
	TotalCount   int `json:"total-count"`
}

// ListOptions is used to specify pagination options when making API requests.
// Pagination allows breaking up large result sets into chunks, or "pages".
type ListOptions struct {
	// The page number to request. The results vary based on the PageSize.
	PageNumber int `schema:"page[number],omitempty"`

	// The number of elements returned in a single page.
	PageSize int `schema:"page[size],omitempty"`
}

// validString checks if the given input is present and non-empty.
func validString(v *string) bool {
	return v != nil && *v != ""
}

// validStringID checks if the given string pointer is non-nil and
// contains a typical string identifier.
func validStringID(v *string) bool {
	return v != nil && reStringID.MatchString(*v)
}

// validStringID checks if the given string pointer is non-nil and contains a
// valid semantic version (major.minor.patch).
func validSemanticVersion(v string) bool {
	return reSemanticVersion.MatchString(v)
}
