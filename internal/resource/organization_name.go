package resource

import (
	"database/sql/driver"
	"fmt"

	"github.com/leg100/otf/internal"
)

type OrganizationName struct {
	name string
}

func NewOrganizationName(name string) (OrganizationName, error) {
	if !validName.MatchString(name) {
		return OrganizationName{}, internal.ErrInvalidName
	}
	return OrganizationName{name: name}, nil
}

func (n OrganizationName) GetKind() Kind  { return OrganizationKind }
func (n OrganizationName) String() string { return n.name }

func (name OrganizationName) MarshalText() ([]byte, error) {
	return []byte(name.name), nil
}

func (name *OrganizationName) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	s := string(text)
	name.name = s
	return nil
}

func (name *OrganizationName) Scan(text any) error {
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

func (name *OrganizationName) Value() (driver.Value, error) {
	if name == nil {
		return nil, nil
	}
	return name.name, nil
}
