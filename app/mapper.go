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
}
