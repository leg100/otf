/*
Package otf is responsible for domain logic.
*/
package otf

import (
	"context"
	crypto "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/sql/pggen"
)

const (
	alphanumeric = "abcdefghijkmnopqrstuvwxyzABCDEFGHIJKMNOPQRSTUVWXYZ0123456789"

	// ChunkStartMarker is the special byte that prefixes the first chunk
	ChunkStartMarker = byte(2)

	// ChunkEndMarker is the special byte that suffixes the last chunk
	ChunkEndMarker = byte(3)
)

var (
	// A regular expression used to validate common string ID patterns.
	reStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)

	// A regular expression used to validate semantic versions (major.minor.patch).
	reSemanticVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
)

// Application provides access to the otf application services
type Application interface {
	// Tx provides a transaction within which to operate on the store.
	Tx(ctx context.Context, tx func(Application) error) error
	DB() DB
	OrganizationService
	WorkspaceService
	StateVersionService
	ConfigurationVersionService
	RunService
	EventService
	UserService
	SessionService
	TokenService
	TeamService
	AgentTokenService
	CurrentRunService
	VCSProviderService
	LockableApplication
	cloud.Service
	ModuleService
	ModuleVersionService
	HostnameService
	HookService
}

// LockableApplication is an application that holds an exclusive lock with the given ID.
type LockableApplication interface {
	WithLock(ctx context.Context, id int64, cb func(Application) error) error
}

// DB provides access to otf database
type DB interface {
	Database

	Tx(ctx context.Context, tx func(DB) error) error
	// WaitAndLock obtains a DB with a session-level advisory lock.
	WaitAndLock(ctx context.Context, id int64, cb func(DB) error) error
	Close()
	UserStore
	TeamStore
	OrganizationStore
	WorkspaceStore
	RunStore
	SessionStore
	TokenStore
	ConfigurationVersionStore
	ChunkStore
	AgentTokenStore
	VCSProviderStore
	ModuleStore
	ModuleVersionStore
}

// Database provides access to generated SQL queries as well as wrappers for
// performing queries within a transaction or a lock.
type Database interface {
	// Send batches of SQL queries over the wire.
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults

	pggen.Querier // generated SQL queries

	// Tx provides a transaction within which to operate on the store.
	Transaction(ctx context.Context, tx func(Database) error) error
	// WaitAndLock obtains a DB with a session-level advisory lock.
	WaitAndLock(ctx context.Context, id int64, cb func(DB) error) error
}

// Unmarshaler unmarshals database rows
type Unmarshaler struct {
	cloud.Service
}

// Identity is an identifiable otf entity.
type Identity interface {
	// Human friendly identification of the entity.
	String() string
	// Uniquely identifies the entity.
	ID() string
}

func String(str string) *string   { return &str }
func Int(i int) *int              { return &i }
func Int64(i int64) *int64        { return &i }
func UInt(i uint) *uint           { return &i }
func Bool(b bool) *bool           { return &b }
func Time(t time.Time) *time.Time { return &t }
func UUID(u uuid.UUID) *uuid.UUID { return &u }

// CurrentTimestamp is *the* way to get a current timestamps in otf and
// time.Now() should be avoided.
//
// We want timestamps to be rounded to nearest
// millisecond so that they can be persisted/serialised and not lose precision
// thereby making comparisons and testing easier.
//
// We also want timestamps to be in the UTC time zone. Again it makes
// testing easier because libs such as testify's assert use DeepEqual rather
// than time.Equal to compare times (and structs containing times). That means
// the internal representation is compared, including the time zone which may
// differ even though two times refer to the same instant.
//
// In any case, the time zone of the server is often not of importance, whereas
// that of the user often is, and conversion to their time zone is necessary
// regardless.
func CurrentTimestamp() time.Time {
	return time.Now().Round(time.Millisecond).UTC()
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
	rand.New(rand.NewSource(time.Now().UnixNano()))

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

func (r ResourceReport) String() string {
	// \u2212 is a proper minus sign; an ascii hyphen is too narrow (in the
	// default github font at least) and looks incongruous alongside
	// the wider '+' and '~' characters.
	return fmt.Sprintf("+%d/~%d/\u2212%d", r.Additions, r.Changes, r.Destructions)
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

// GenerateAuthToken generates an authentication token for a type of account
// e.g. agent, user
func GenerateAuthToken(accountType string) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}
	return accountType + "." + token, nil
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

// Exists checks whether a file or directory at the given path exists
func Exists(path string) bool {
	// Interpret any error from os.Stat as "not found"
	_, err := os.Stat(path)
	return err == nil
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
func Index[E comparable](s []E, v E) int {
	for i, vs := range s {
		if v == vs {
			return i
		}
	}
	return -1
}

// Contains reports whether v is present in s.
func Contains[E comparable](s []E, v E) bool {
	return Index(s, v) >= 0
}
