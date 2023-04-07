// Package integration provides inter-service integration tests.
package integration

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
)

// Context conferring site admin privileges
var ctx = otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)
