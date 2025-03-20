package resource

import (
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/require"
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

func (OrganizationName) GetKind() Kind       { return OrganizationKind }
func (name OrganizationName) String() string { return name.name }

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

// Value satisfies the pflag.Value interface
func (name *OrganizationName) Value() (driver.Value, error) {
	if name == nil {
		return nil, nil
	}
	return name.name, nil
}

// Set satisfies the pflag.Value interface
func (name *OrganizationName) Set(v string) error {
	validated, err := NewOrganizationName(v)
	if err != nil {
		return err
	}
	*name = validated
	return nil
}

// Type satisfies the pflag.Value interface
func (name *OrganizationName) Type() string {
	return "organization"
}

func NewTestOrganizationName(t *testing.T) OrganizationName {
	name, err := NewOrganizationName(uuid.NewString())
	require.NoError(t, err)
	return name
}
