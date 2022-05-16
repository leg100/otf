/*
Package otf is responsible for domain logic.
*/
package otf

import (
	crypto "crypto/rand"
	"encoding/base64"
	"math"
	"math/rand"
	"regexp"
	"time"

	"github.com/jackc/pgtype"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 100

	DefaultUserID   = "user-123"
	DefaultUsername = "otf"

	alphanumeric = "abcdefghijkmnopqrstuvwxyzABCDEFGHIJKMNOPQRSTUVWXYZ0123456789"

	// ChunkMaxLimit is maximum permissible size of a chunk
	ChunkMaxLimit = 65536

	// ChunkStartMarker is the special byte that prefixes the first chunk
	ChunkStartMarker = byte(2)

	// ChunkEndMarker is the special byte that suffixes the last chunk
	ChunkEndMarker = byte(3)
)

// A regular expression used to validate common string ID patterns.
var reStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)

// A regular expression used to validate semantic versions (major.minor.patch).
var reSemanticVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// Application provides access to the oTF application services
type Application interface {
	OrganizationService() OrganizationService
	WorkspaceService() WorkspaceService
	StateVersionService() StateVersionService
	ConfigurationVersionService() ConfigurationVersionService
	RunService() RunService
	PlanService() PlanService
	ApplyService() ApplyService
	EventService() EventService
	UserService() UserService
	//GetCacheService() *CacheService
}

// DB provides access to oTF database
type DB interface {
	Close() error

	OrganizationStore() OrganizationStore
	WorkspaceStore() WorkspaceStore
	StateVersionStore() StateVersionStore
	ConfigurationVersionStore() ConfigurationVersionStore
	RunStore() RunStore
	PlanLogStore() PlanLogStore
	ApplyLogStore() ApplyLogStore
	UserStore() UserStore
	SessionStore() SessionStore
	TokenStore() TokenStore
}

// Updateable is an obj that records when it was updated.
type Updateable interface {
	GetID() string
	SetUpdatedAt(time.Time)
}

func String(str string) *string { return &str }
func Int(i int) *int            { return &i }
func Int64(i int64) *int64      { return &i }
func UInt(i uint) *uint         { return &i }

// TimeNow is a convenience func to return the pointer of the current time
func TimeNow() *time.Time {
	t := time.Now()
	return &t
}

// NewID constructs resource IDs, which are composed of the resource type and a
// random 16 character string, separated by a hyphen.
func NewID(rtype string) string {
	return rtype + "-" + GenerateRandomString(16)
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

var _ pgtype.BinaryDecoder = (*ResourceReport)(nil)

// ResourceReport reports a summary of additions, changes, and deletions of
// resources in a plan or an apply.
type ResourceReport struct {
	Additions    int `json:"additions"`
	Changes      int `json:"changes"`
	Destructions int `json:"destructions"`
}

func (t *ResourceReport) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	r := pgtype.Record{
		Fields: []pgtype.Value{&pgtype.Int4{}, &pgtype.Int4{}, &pgtype.Int4{}},
	}

	if err := r.DecodeBinary(ci, src); err != nil {
		return err
	}

	// NULL -> nil
	if r.Status != pgtype.Present {
		t = nil
		return nil
	}

	a := r.Fields[0].(*pgtype.Int4)
	b := r.Fields[1].(*pgtype.Int4)
	c := r.Fields[2].(*pgtype.Int4)

	// type compatibility is checked by AssignTo
	// only lossless assignments will succeed
	if err := a.AssignTo(&t.Additions); err != nil {
		return err
	}

	// AssignTo also deals with null value handling
	if err := b.AssignTo(&t.Changes); err != nil {
		return err
	}

	// AssignTo also deals with null value handling
	if err := c.AssignTo(&t.Destructions); err != nil {
		return err
	}
	return nil
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

// GetOffset calculates the offset for use in SQL queries.
func (o *ListOptions) GetOffset() int {
	if o.PageNumber == 0 {
		return 0
	}

	return (o.PageNumber - 1) * o.PageSize
}

// GetLimit calculates the limit for use in SQL queries.
func (o *ListOptions) GetLimit() int {
	// TODO: remove MaxPageSize - this is too complicated
	if o.PageSize == 0 {
		return math.MaxInt
	} else if o.PageSize > MaxPageSize {
		return MaxPageSize
	}

	return o.PageSize
}

// validString checks if the given input is present and non-empty.
func validString(v *string) bool {
	return v != nil && *v != ""
}

// ValidStringID checks if the given string pointer is non-nil and
// contains a typical string identifier.
func ValidStringID(v *string) bool {
	return v != nil && reStringID.MatchString(*v)
}

// validStringID checks if the given string pointer is non-nil and contains a
// valid semantic version (major.minor.patch).
func validSemanticVersion(v string) bool {
	return reSemanticVersion.MatchString(v)
}

type Timestamps struct {
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (m *Timestamps) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = t
}

func NewTimestamps() Timestamps {
	now := time.Now()
	return Timestamps{
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func GetMapKeys(m map[string]interface{}) []string {
	keys := make([]string, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// PrefixSlice prefixes each string in a slice with another string.
func PrefixSlice(slice []string, prefix string) (ret []string) {
	for _, s := range slice {
		ret = append(ret, prefix+s)
	}
	return
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := crypto.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
