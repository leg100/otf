package otf

import "github.com/google/uuid"

// WorkspaceRepo represents a connection between a workspace and a VCS
// repository.
//
// TODO: rename WorkspaceConnection
type WorkspaceRepo struct {
	ProviderID string
	WebhookID  uuid.UUID
	Identifier string // identifier is <repo_owner>/<repo_name>
	Branch     string // branch for which applies are run
}

func NewWorkspaceRepo(opts NewWorkspaceRepoOptions) WorkspaceRepo {
	return WorkspaceRepo{
		ProviderID: opts.ProviderID,
		WebhookID:  opts.WebhookID,
		Identifier: opts.Identifier,
		Branch:     opts.Branch,
	}
}

type NewWorkspaceRepoOptions struct {
	ProviderID string
	Branch     string
	*Webhook
}

// WorkspaceUpdateRepoOptions are options for updating a workspace's associated
// repo.
type WorkspaceUpdateRepoOptions struct {
	Branch *string
}

// unmarshal workspace repo
