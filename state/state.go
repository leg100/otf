// Package state manages terraform state.
package state

import (
	"github.com/leg100/otf/workspace"
)

type (
	CreateStateVersionOptions struct {
		State       []byte  // Terraform state file. Required.
		WorkspaceID *string // ID of state version's workspace. Required.
		Serial      *int64  // State serial number. If not provided then it is extracted from the state.
	}

	// Alias services so they don't conflict when nested together in struct
	WorkspaceService workspace.Service
)
