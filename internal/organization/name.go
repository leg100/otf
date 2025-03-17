package organization

import (
	"database/sql/driver"
	"fmt"
	"regexp"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

type Name struct {
	name string
}

func NewName(name string) (Name, error) {
	if !validName.MatchString(name) {
		return Name{}, internal.ErrInvalidName
	}
	return Name{name: name}, nil
}

func (n Name) GetKind() resource.Kind { return resource.OrganizationKind }
func (n Name) String() string         { return n.name }

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

func (name *Name) Value() (driver.Value, error) {
	if name == nil {
		return nil, nil
	}
	return name.name, nil
}
