package resource

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
)

const (
	// base58 alphabet
	base58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	// length of id part of ID
	idLength = 16
)

var (
	EmptyID = ID{}
	// ReStringID is a regular expression used to validate common string ID patterns.
	ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
	// regex for valid ID
	idRegex = regexp.MustCompile(`^[a-z]{2,}-[` + base58 + `]{` + strconv.Itoa(idLength) + `}`)
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

func ParseID(s string) (ID, error) {
	matches := idRegex.FindStringSubmatch(s)
	if matches == nil {
		return ID{}, fmt.Errorf("failed to parse resource ID: %s", s)
	}
	return ID{Kind: Kind(matches[1]), ID: matches[2]}, nil
}

func (id ID) String() string {
	return fmt.Sprintf("%s-%s", id.Kind, id.ID)
}

func (id *ID) UnmarshalText(text []byte) error {
	s := string(text)
	x, err := ParseID(s)
	if err != nil {
		return fmt.Errorf("failed to parse resource ID: %s", s)
	}
	*id = x
	return nil
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
