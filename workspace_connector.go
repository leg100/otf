package otf

import (
	"context"
)

// WorkspaceConnector connects a workspace to a VCS repo, subscribing it to
// VCS events that trigger runs.
type WorkspaceConnector interface {
	Connect(ctx context.Context, workspaceID string, opts ConnectWorkspaceOptions) error
	Disconnect(ctx context.Context, workspaceID string) (*Workspace, error)
}

type ConnectWorkspaceOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
}
