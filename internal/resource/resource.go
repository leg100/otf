// Package resource contains code common to all resources (orgs, workspaces,
// runs, etc)
package resource

import (
	"regexp"

	"github.com/leg100/otf/internal"
)

// A regular expression used to validate resource name.
var validName = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

func ValidateName(name *string) error {
	if name == nil {
		return internal.ErrRequiredName
	}
	if !validName.MatchString(*name) {
		return internal.ErrInvalidName
	}
	return nil
}
