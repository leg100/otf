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
	// ErrInvalidTerraformVersion is returned when a terraform version string is
	// not a semantic version string (major.minor.patch).
	ErrInvalidTerraformVersion = errors.New("invalid terraform version")

	// ErrRequiredOrg is returned when the organization option is not present
	ErrRequiredOrg = errors.New("organization is required")

	ErrStatusTimestampNotFound = errors.New("corresponding status timestamp not found")

	ErrInvalidRepo = errors.New("repository path is invalid")
)

type (
	HTTPError struct {
		Code    int
		Message string
	}

	// MissingParameterError occurs when the caller has failed to provide a
	// required parameter
	MissingParameterError struct {
		Parameter string
	}

	// ForeignKeyError occurs when there is a foreign key violation.
	ForeignKeyError struct {
		*pgconn.PgError
	}

	InvalidParameterError string
)

func (e InvalidParameterError) Error() string {
	return string(e)
}

func (e *HTTPError) Error() string {
	return e.Message
}

func (e *MissingParameterError) Error() string {
	return fmt.Sprintf("required parameter missing: %s", e.Parameter)
}

func (e *ForeignKeyError) Error() string {
	return e.Detail
}
