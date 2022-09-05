/*
Package otf is responsible for domain logic.
*/
package otf

import (
	"context"
	crypto "crypto/rand"
	"encoding/base64"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const (
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
	OrganizationService
	WorkspaceService
	StateVersionService
	ConfigurationVersionService
	RunService
	EventService
	UserService
	AgentTokenService
	LatestRunService
}

// DB provides access to oTF database
type DB interface {
	// Tx provides a transaction within which to operate on the store.
	Tx(ctx context.Context, tx func(DB) error) error
	Close()
	UserStore
	OrganizationStore
	WorkspaceStore
	RunStore
	SessionStore
	StateVersionStore
	TokenStore
	ConfigurationVersionStore
	ChunkStore
	AgentTokenStore
}

// Identity is an identifiable oTF entity.
type Identity interface {
	// Human friendly identification of the entity.
	String() string
	// Uniquely identifies the entity.
	ID() string
}

func String(str string) *string { return &str }
func Int(i int) *int            { return &i }
func Int64(i int64) *int64      { return &i }
func UInt(i uint) *uint         { return &i }
func Bool(b bool) *bool         { return &b }

// CurrentTimestamp is *the* way to get a current timestamps in oTF and
// time.Now() should be avoided. We want timestamps to be rounded to nearest
// millisecond so that they can be persisted/serialised and not lose precision
// thereby making comparisons and testing easier.
func CurrentTimestamp() time.Time {
	return time.Now().Round(time.Millisecond)
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

// ResourceReport reports a summary of additions, changes, and deletions of
// resources in a plan or an apply.
type ResourceReport struct {
	Additions    int
	Changes      int
	Destructions int
}

func (r ResourceReport) HasChanges() bool {
	if r.Additions > 0 || r.Changes > 0 || r.Destructions > 0 {
		return true
	}
	return false
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

// ConvertID converts an ID for use with a different resource, e.g. convert
// run-123 to plan-123.
func ConvertID(id, resource string) string {
	parts := strings.Split(id, "-")
	// if ID not in expected form then just return it unchanged without error
	if len(parts) != 2 {
		return id
	}
	return resource + "-" + parts[1]
}

// AppUser identifies the otf app itself for purposes of authentication. Some
// processes require more privileged access than the invoking user possesses, so
// it is necessary to escalate privileges by "sudo'ing" to this user.
type AppUser struct{}

func (*AppUser) CanAccess(*string) bool { return true }
func (*AppUser) String() string         { return "app-user" }
func (*AppUser) ID() string             { return "app-user" }

func ContextWithAppUser() context.Context {
	return AddSubjectToContext(context.Background(), &AppUser{})
}
