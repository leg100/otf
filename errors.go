package otf

import (
	"errors"

	"gorm.io/gorm"
)

// Generic errors applicable to all resources.
var (
	// ErrRequiredName is returned when a name option is not present.
	ErrRequiredName = errors.New("name is required")

	// ErrInvalidName is returned when the name option has invalid value.
	ErrInvalidName = errors.New("invalid value for name")
)

// Resource Errors
var (
	// ErrInvalidTerraformVersion is returned when a terraform version string is
	// not a semantic version string (major.minor.patch).
	ErrInvalidTerraformVersion = errors.New("invalid terraform version")
)

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
