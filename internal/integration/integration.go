// Package integration provides inter-service integration tests.
package integration

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
)

var (
	// shared environment variables for individual tests to use
	envs []string

	// Context conferring site admin privileges
	ctx = internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)
)
