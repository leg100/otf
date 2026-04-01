package policy

import (
	"database/sql/driver"
	"fmt"

	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/semver"
)

// Version is a policy engine version.
type Version struct {
	Latest bool
	semver string
}

func LatestVersion() Version {
	return Version{Latest: true}
}

func (v Version) String() string {
	if v.Latest {
		return "latest"
	}
	return v.semver
}

func (v Version) MarshalText() ([]byte, error) {
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

func (v Version) Value() (driver.Value, error) {
	return v.String(), nil
}

func (v *Version) set(value string) error {
	switch value {
	case "latest":
		v.Latest = true
		v.semver = ""
		return nil
	case "":
		return fmt.Errorf("value cannot be an empty string")
	default:
		if !semver.IsValid(value) {
			return engine.ErrInvalidVersion
		}
		v.Latest = false
		v.semver = value
		return nil
	}
}

func (k Kind) Valid() bool {
	switch k {
	case SentinelKind, OPAKind:
		return true
	default:
		return false
	}
}
