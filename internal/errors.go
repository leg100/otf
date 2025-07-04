package internal

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// Generic errors
var (
	// ErrAccessNotPermitted is returned when an authorization check fails.
	ErrAccessNotPermitted = errors.New("access to the resource is not permitted")

	// ErrUnauthorized is returned when a receiving a 401.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrResourceNotFound is returned when a receiving a 404.
	ErrResourceNotFound = errors.New("resource not found")

	// ErrResourceAlreadyExists is returned when attempting to create a resource
	// that already exists.
	ErrResourceAlreadyExists = errors.New("resource already exists")

	// ErrRequiredName is returned when a name option is not present.
	ErrRequiredName = errors.New("name is required")

	// ErrInvalidName is returned when the name option has invalid value.
	ErrInvalidName = errors.New("invalid value for name")

	// ErrEmptyValue is returned when a value is set to an empty string
	ErrEmptyValue = errors.New("value cannot be empty")

	// ErrTimeout is returned when a request exceeds a timeout.
	ErrTimeout = errors.New("request timed out")

	// ErrConflict is returned when a requests attempts to either create a
	// resource with an identifier that already exists, or if an invalid state
	// transition is attempted
	ErrConflict = errors.New("resource conflict detected")
)

// Resource Errors
var (
	// ErrRequiredOrg is returned when the organization option is not present
	ErrRequiredOrg = errors.New("organization is required")

	ErrStatusTimestampNotFound = errors.New("corresponding status timestamp not found")
)

// ForeignKeyError occurs when there is a foreign key violation.
type ForeignKeyError struct {
	*pgconn.PgError
}

func (e *ForeignKeyError) Error() string {
	return e.Detail
}

// ErrMissingParameter occurs when the user has failed to provide a
// required parameter
type ErrMissingParameter struct {
	Parameter string
}

func (e *ErrMissingParameter) Error() string {
	return fmt.Sprintf("required parameter missing: %s", e.Parameter)
}

// ErrorIs is a modification to the upstream errors.Is, allowing multiple
// targets to be checked.
func ErrorIs(err error, target error, moreTargets ...error) bool {
	for _, t := range append(moreTargets, target) {
		if errors.Is(err, t) {
			return true
		}
	}
	return false
}
