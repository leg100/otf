/*
Package otf is responsible for domain logic.
*/
package otf

import (
	"context"
	crypto "crypto/rand"
	"encoding/base64"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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
	ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)

	// A regular expression used to validate semantic versions (major.minor.patch).
	ReSemanticVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
)

// DB is the otf database. Services may wrap this and implement higher-level
// queries.
type DB interface {
	// Tx provides a callback in which queries are run within a transaction.
	Tx(ctx context.Context, tx func(DB) error) error
	// Acquire dedicated connection from connection pool.
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	// Execute arbitrary SQL
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	// Send batches of SQL queries over the wire.
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults

	pggen.Querier // queries generated from SQL
	Close()       // Close all connections in pool

	// additional queries that wrap the generated queries
	GetLogs(ctx context.Context, runID string, phase PhaseType) ([]byte, error)
}

type DatabaseLock interface {
	Release()
}

// GetID retrieves the ID field of a struct contained in s. If s is not a struct,
// or there is no ID field, then false is returned.
func GetID(s any) (string, bool) {
	v := reflect.Indirect(reflect.ValueOf(s))
	if v.Kind() != reflect.Struct {
		return "", false
	}
	f := v.FieldByName("ID")
	if !f.IsValid() {
		return "", false
	}
	return f.String(), true
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

// ValidStringID checks if the given string pointer is non-nil and
// contains a typical string identifier.
func ValidStringID(v *string) bool {
	return v != nil && ReStringID.MatchString(*v)
}

// ValidSemanticVersion checks if v is a
// valid semantic version (major.minor.patch).
func ValidSemanticVersion(v string) bool {
	return ReSemanticVersion.MatchString(v)
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
