package resource

import (
	"database/sql/driver"
	"fmt"
	"math/rand"
	"strings"
)

// base58 alphabet
const base58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var _ ID = (*TfeID)(nil)

// TfeID is an ID using the same format used for Terraform Enterprise resources,
// with a prefix designating the type of resource, appended with a unique base58
// encoded id.
type TfeID struct {
	kind Kind
	id   string
}

// NewTfeID constructs a resource ID
func NewTfeID(kind Kind) TfeID {
	return TfeID{kind: kind, id: GenerateRandomStringFromAlphabet(16, base58)}
}

// ConvertTfeID converts an ID for use with a different resource kind, e.g. convert
// run-123 to plan-123.
func ConvertTfeID(id TfeID, to Kind) TfeID {
	return TfeID{kind: to, id: id.id}
}

func MustHardcodeTfeID(kind Kind, suffix string) TfeID {
	s := fmt.Sprintf("%s-%s", kind, suffix)
	id, err := ParseTfeID(s)
	if err != nil {
		panic("failed to parse hardcoded ID: " + err.Error())
	}
	return id
}

// ParseTfeID parses the ID from a string representation.
func ParseTfeID(s string) (TfeID, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return TfeID{}, fmt.Errorf("malformed ID: %s", s)
	}
	kind := parts[0]
	if len(kind) < 2 {
		return TfeID{}, fmt.Errorf("kind must be at least 2 characters: %s", s)
	}
	id := parts[1]
	if len(id) < 1 {
		return TfeID{}, fmt.Errorf("id suffix must be at least 1 character: %s", s)
	}
	return TfeID{kind: Kind(kind), id: id}, nil
}

func (id TfeID) String() string {
	return fmt.Sprintf("%s-%s", id.kind, id.id)
}

func (id TfeID) IsZero() bool {
	return id.id == "" && id.kind == ""
}

func (id TfeID) Kind() Kind {
	return id.kind
}

func (id TfeID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *TfeID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	s := string(text)
	x, err := ParseTfeID(s)
	if err != nil {
		return err
	}
	*id = x
	return nil
}

func (id *TfeID) Scan(text any) error {
	if text == nil {
		return nil
	}
	s, ok := text.(string)
	if !ok {
		return fmt.Errorf("expected database value to be a string: %#v", text)
	}
	x, err := ParseTfeID(s)
	if err != nil {
		return err
	}
	*id = x
	return nil
}

func (id *TfeID) Value() (driver.Value, error) {
	if id == nil {
		return nil, nil
	}
	return id.String(), nil
}

// Type implements pflag.Value
func (*TfeID) Type() string { return "id" }

// Set implements pflag.Value
func (id *TfeID) Set(text string) error { return id.Scan(text) }

// GenerateRandomStringFromAlphabet generates a random string of a given size
// using characters from the given alphabet.
func GenerateRandomStringFromAlphabet(size int, alphabet string) string {
	buf := make([]byte, size)
	for i := range size {
		buf[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(buf)
}
