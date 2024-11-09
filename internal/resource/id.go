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

// ParseID parses the ID from a string representation.
func ParseID(s string) (ID, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return ID{}, fmt.Errorf("malformed ID: %s", s)
	}
	kind := parts[0]
	id := parts[1]
	return ID{Kind: Kind(kind), ID: id}, nil
}

func (id ID) String() string {
	return fmt.Sprintf("%s-%s", id.Kind, id.ID)
}

func (id *ID) UnmarshalText(text []byte) error {
	// string also makes a copy which is necessary in order to retain the data
	// after returning.
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
