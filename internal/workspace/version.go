package workspace

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"slices"

	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/semver"
)

var apiTestTerraformVersions = []string{"0.10.0", "0.11.0", "0.11.1"}

// Version is a workspace's engine version.
type Version struct {
	// Latest if true means runs use the Latest available version at time of
	// creation of the run.
	Latest bool
	// semver is the semantic version of the engine; must be non-empty if latest
	// is false.
	//
	// TODO: use custom type
	semver string
}

func (v *Version) String() string {
	if v.Latest {
		return "latest"
	}
	return v.semver
}

func (v *Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Version) UnmarshalText(text []byte) error {
	return v.set(string(text))
}

func (v *Version) Scan(text any) error {
	if text == nil {
		return nil
	}
	s, ok := text.(string)
	if !ok {
		return fmt.Errorf("expected database value to be a string: %#v", text)
	}
	return v.set(s)
}

func (v *Version) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return v.String(), nil
}

var errEmptyString = errors.New("value cannot be an empty string")

func (v *Version) set(value string) error {
	switch value {
	case "latest":
		v.Latest = true
	case "":
		return errEmptyString
	default:
		if !semver.IsValid(value) {
			return engine.ErrInvalidVersion
		}
		// only accept engine versions above the minimum requirement.
		//
		// NOTE: we make an exception for the specific versions posted by the go-tfe
		// integration tests.
		if result := semver.Compare(value, engine.MinEngineVersion); result < 0 {
			if !slices.Contains(apiTestTerraformVersions, value) {
				return ErrUnsupportedTerraformVersion
			}
		}
		v.semver = value
	}
	return nil
}
