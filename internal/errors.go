package internal

import (
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
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

	// ErrUploadTooLarge is returned when a user attempts to upload data that
	// is too large.
	ErrUploadTooLarge = errors.New("upload is too large")
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

// Workspace errors
var (
	ErrWorkspaceAlreadyLocked         = errors.New("workspace already locked")
	ErrWorkspaceLockedByDifferentUser = errors.New("workspace locked by different user")
	ErrWorkspaceLockedByRun           = errors.New("workspace is locked by Run")
	ErrWorkspaceAlreadyUnlocked       = errors.New("workspace already unlocked")
	ErrWorkspaceUnlockDenied          = errors.New("unauthorized to unlock workspace")
	ErrWorkspaceInvalidLock           = errors.New("invalid workspace lock")
	ErrUnsupportedTerraformVersion    = errors.New("unsupported terraform version")
)

// Run errors
var (
	ErrRunDiscardNotAllowed     = errors.New("run was not paused for confirmation or priority; discard not allowed")
	ErrRunCancelNotAllowed      = errors.New("run was not planning or applying; cancel not allowed")
	ErrRunForceCancelNotAllowed = errors.New("run was not planning or applying, has not been canceled non-forcefully, or the cool-off period has not yet passed")
	//
	ErrPhaseAlreadyStarted = errors.New("phase already started")
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
