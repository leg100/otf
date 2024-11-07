package resource

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

const (
	// base58 alphabet
	base58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	// length of id part of ID
	idLength = 16
)

var (
	// EmptyID for use in comparisons to check whether ID has been
	// uninitialized.
	EmptyID = ID{}
	// ReStringID is a regular expression used to validate common string ID patterns.
	ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
	// regex for valid ID
	idRegex = regexp.MustCompile(`^[a-z]{2,}-[` + base58 + `]{` + strconv.Itoa(idLength) + `}`)
)

// ID uniquely identifies an OTF resource.
type ID struct {
	Kind Kind
	ID   string
}

// NewID constructs a resource ID
func NewID(kind Kind) ID {
	return ID{Kind: kind, ID: GenerateRandomStringFromAlphabet(16, base58)}
}

// ConvertID converts an ID for use with a different resource kind, e.g. convert
// run-123 to plan-123.
func ConvertID(id ID, to Kind) ID {
	return ID{Kind: to, ID: id.ID}
}

// ParseID parses the ID from a string representation. No validation is
// performed.
//
// TODO(@leg100): perform validation and change signature to return error when
// validation fails. I'm hesistant to do this just yet because this function is
// used heavily to both unmarshal IDs from the DB and in tests, and it's a PITA
// to check errors every time. It might be better to find a way of implementing
// the database/sql.Scan interface, and getting that to work with sqlc; or to
// wait until IDs are migrated over to use UUIDs, which would change a lot of
// things...
func ParseID(s string) ID {
	kind, id, _ := strings.Cut(s, "-")
	return ID{Kind: Kind(kind), ID: id}
}

func IDPtr(id ID) *ID { return &id }

func (id ID) String() string {
	return fmt.Sprintf("%s-%s", id.Kind, id.ID)
}

func (id *ID) UnmarshalText(text []byte) error {
	s := string(text)
	*id = ParseID(s)
	return nil
}

// GetID allows the user of an interface to retrieve the ID.
func (id ID) GetID() ID {
	return id
}

// GenerateRandomStringFromAlphabet generates a random string of a given size
// using characters from the given alphabet.
func GenerateRandomStringFromAlphabet(size int, alphabet string) string {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(buf)
}
