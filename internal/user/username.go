package user

import (
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/require"
)

var (
	_ resource.ID = (*Username)(nil)
)

type Username struct {
	name string
}

func NewUsername(name string) (Username, error) {
	if len(name) == 0 {
		return Username{}, fmt.Errorf("%w: username must be non-empty", internal.ErrInvalidName)
	}
	return Username{name: name}, nil
}

func MustUsername(name string) Username {
	username, err := NewUsername(name)
	if err != nil {
		panic("error creating username: " + err.Error())
	}
	return username
}

func (Username) Kind() resource.Kind { return resource.UserKind }
func (name Username) String() string { return name.name }

func (name Username) MarshalText() ([]byte, error) {
	return []byte(name.name), nil
}

func (name *Username) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	s := string(text)
	name.name = s
	return nil
}

func (name *Username) Scan(text any) error {
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

func (name *Username) Value() (driver.Value, error) {
	if name == nil {
		return nil, nil
	}
	return name.name, nil
}

func NewTestUsername(t *testing.T) Username {
	name, err := NewUsername(uuid.NewString())
	require.NoError(t, err)
	return name
}
