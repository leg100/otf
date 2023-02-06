package workspace

import (
	"github.com/google/uuid"
)

// WorkspaceRepo represents a connection between a workspace and a VCS
// repository.
//
// TODO: rename WorkspaceConnection
type WorkspaceRepo struct {
	ProviderID  string
	WebhookID   uuid.UUID
	Identifier  string // identifier is <repo_owner>/<repo_name>
	Branch      string // branch for which applies are run
	WorkspaceID string
}
