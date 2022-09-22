package app

import (
	"context"

	"github.com/leg100/otf"
)

// Mapper maintains mappings between various resource identifiers, which are
// used by upstream layers to make decisions and efficiently lookup resources.
type Mapper interface {
	Start(context.Context) error
	LookupWorkspaceID(spec otf.WorkspaceSpec) string

	CanAccessRun(ctx context.Context, runID string) bool
	CanAccessWorkspace(ctx context.Context, spec otf.WorkspaceSpec) bool
}
