package app

import (
	"context"

	"github.com/leg100/otf"
)

// Mapper maintains mappings between various resource identifiers, which are
// used by upstream layers to make decisions and efficiently lookup resources.
type Mapper interface {
	Populate(ctx context.Context, ws otf.WorkspaceService, rs otf.RunService) error

	MapRun(run *otf.Run)
	UnmapRun(run *otf.Run)

	MapWorkspace(ws *otf.Workspace)
	RemapWorkspace(oldName string, ws *otf.Workspace)
	UnmapWorkspace(ws *otf.Workspace)

	LookupWorkspaceID(spec otf.WorkspaceSpec) string

	CanAccessRun(ctx context.Context, runID string) bool
	CanAccessWorkspace(ctx context.Context, spec otf.WorkspaceSpec) bool
}
