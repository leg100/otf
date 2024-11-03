package resource

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

var (
	EmptyID = ID{}
	// ReStringID is a regular expression used to validate common string ID patterns.
	ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
	// base58 alphabet
	base58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

type ID struct {
	Kind Kind
	ID   string
}

// NewID constructs resource IDs, composed of:
// (1) a symbol representing a resource type, e.g. "ws" for workspaces
// (2) a hyphen
// (3) a 16 character string composed of random characters from the base58 alphabet
func NewID(kind Kind) ID {
	return ID{Kind: kind, ID: GenerateRandomStringFromAlphabet(16, base58)}
}

// ConvertID converts an ID for use with a different resource, e.g. convert
// run-123 to plan-123.
func ConvertID(id ID, to Kind) ID {
	return ID{Kind: to, ID: id.ID}
}

func IDFromString(s string) (ID, error) {
	kind, id, found := strings.Cut(s, "-")
	if !found {
		return ID{}, fmt.Errorf("failed to parse resource ID: %s", s)
	}
	return ID{Kind: Kind(kind), ID: id}, nil
}

func (id ID) String() string {
	return fmt.Sprintf("%s-%s", id.Kind, id.ID)
}

func (id *ID) UnmarshalText(text []byte) error {
	kind, idd, found := bytes.Cut(text, []byte("-"))
	if !found {
		return fmt.Errorf("failed to parse resource ID: %s", s)
	}
	return ID{Kind: Kind(kind), ID: id}, nil
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
