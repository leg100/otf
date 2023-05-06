// Package integration provides inter-service integration tests.
package integration

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
)

var (
	// shared environment variables for individual tests to use
	envs []string

	// Context conferring site admin privileges
	ctx = otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)
)
