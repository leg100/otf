package resource

import (
	"database/sql/driver"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

// base58 alphabet
const base58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var (
	// EmptyID for use in comparisons to check whether ID has been
	// uninitialized.
	EmptyID = ID{}
	// ReStringID is a regular expression used to validate common string ID patterns.
	ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
)

// ID uniquely identifies an OTF resource.
type ID struct {
	kind Kind
	id   string
}

// NewID constructs a resource ID
func NewID(kind Kind) ID {
	return ID{kind: kind, id: GenerateRandomStringFromAlphabet(16, base58)}
}

// ConvertID converts an ID for use with a different resource kind, e.g. convert
// run-123 to plan-123.
func ConvertID(id ID, to Kind) ID {
	return ID{kind: to, id: id.id}
}

func MustHardcodeID(kind Kind, suffix string) ID {
	s := fmt.Sprintf("%s-%s", kind, suffix)
	id, err := ParseID(s)
	if err != nil {
		panic("failed to parse hardcoded ID: " + err.Error())
	}
	return id
}

// ParseID parses the ID from a string representation.
func ParseID(s string) (ID, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return ID{}, fmt.Errorf("malformed ID: %s", s)
	}
	kind := parts[0]
	if len(kind) < 2 {
		return ID{}, fmt.Errorf("kind must be at least 2 characters: %s", s)
	}
	id := parts[1]
	if len(id) < 1 {
		return ID{}, fmt.Errorf("id suffix must be at least 1 character: %s", s)
	}
	return ID{kind: Kind(kind), id: id}, nil
}

func (id ID) String() string {
	return fmt.Sprintf("%s-%s", id.kind, id.id)
}

func (id ID) Kind() Kind {
	return id.kind
}

func (id ID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	s := string(text)
	x, err := ParseID(s)
	if err != nil {
		return err
	}
	*id = x
	return nil
}

func (id *ID) Scan(text any) error {
	if text == nil {
		return nil
	}
	s, ok := text.(string)
	if !ok {
		return fmt.Errorf("expected database value to be a string: %#v", text)
	}
	x, err := ParseID(s)
	if err != nil {
		return err
	}
	*id = x
	return nil
}

func (id *ID) Value() (driver.Value, error) {
	if id == nil {
		return nil, nil
	}
	return id.String(), nil
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
