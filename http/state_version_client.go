package http

import "github.com/leg100/otf"

// Compile-time proof of interface implementation.
var _ otf.StateVersionService = (*stateVersions)(nil)

// stateVersions implements StateVersionService.
type stateVersions struct {
	client Client

	// TODO: implement all of otf.StateVersionService's methods
	otf.StateVersionService
}
