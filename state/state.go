// Package state manages terraform state.
package state

import (
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/workspace"
)

type (
	// Alias services so they don't conflict when nested together in struct
	ConfigurationVersionService configversion.Service
	WorkspaceService            workspace.Service
)
