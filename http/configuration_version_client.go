package http

import "github.com/leg100/otf"

// Compile-time proof of interface implementation.
var _ otf.ConfigurationVersionService = (*configurationVersions)(nil)

// configurationVersions implements ConfigurationVersionService.
type configurationVersions struct {
	client Client
	// TODO: implement all of otf.ConfigurationVersionService's methods
	otf.ConfigurationVersionService
}
