package organization

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/require"
)

var (
	_         resource.ID = (*Name)(nil)
	validName             = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
)

type Name struct {
	name string
}

func NewName(name string) (Name, error) {
	if !validName.MatchString(name) {
		return Name{}, internal.ErrInvalidName
	}
	return Name{name: name}, nil
}

func (Name) Kind() resource.Kind { return resource.OrganizationKind }
func (name Name) String() string { return name.name }

func (name Name) MarshalText() ([]byte, error) {
	return []byte(name.name), nil
}

func (name *Name) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	s := string(text)
	name.name = s
	return nil
}

func (name *Name) Scan(text any) error {
	if text == nil {
		return nil
	}
	s, ok := text.(string)
	if !ok {
		return fmt.Errorf("expected database value to be a string: %#v", text)
	}
	name.name = s
	return nil
}

// Value satisfies the pflag.Value interface
func (name *Name) Value() (driver.Value, error) {
	if name == nil {
		return nil, nil
	}
	return name.name, nil
}

// Set satisfies the pflag.Value interface
func (name *Name) Set(v string) error {
	validated, err := NewName(v)
	if err != nil {
		return err
	}
	*name = validated
	return nil
}

// Type satisfies the pflag.Value interface
func (name *Name) Type() string {
	return "organization"
}

func NewTestName(t *testing.T) Name {
	name, err := NewName(uuid.NewString())
	require.NoError(t, err)
	return name
}
