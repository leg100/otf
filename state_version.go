package otf

import (
	"context"

	"github.com/gorilla/mux"
)

// StateVersionService provides to the state app as well as registering handlers
// for access via http
type StateVersionService interface {
	// AddHandlers adds http handlers for to the given mux. The handlers
	// implement the state service API.
	AddHandlers(*mux.Router)

	StateVersionApp
}

// StateVersionApp provides access to creating and retrieving state.
type StateVersionApp interface {
	// CreateStateVersion creates a state version for the given workspace using
	// the given state data.
	CreateStateVersion(ctx context.Context, opts CreateStateVersionOptions) error
	// DownloadCurrentState downloads the current (latest) state for the given
	// workspace.
	DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error)
}

type CreateStateVersionOptions struct {
	State       []byte  // Terraform state file. Required.
	WorkspaceID *string // ID of state version's workspace. Required.
}
