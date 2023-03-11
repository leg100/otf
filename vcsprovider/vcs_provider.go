// Package vcsprovider is responsible for VCS providers
package vcsprovider

import (
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/workspace"
)

type (
	// Alias services so they don't conflict when nested together in struct
	CloudService                cloud.Service
	ConfigurationVersionService configversion.Service
	WorkspaceService            workspace.Service
)
