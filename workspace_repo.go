package otf

// WorkspaceRepo represents a connection between a workspace and a VCS
// repository.
//
// TODO: rename WorkspaceConnection
type WorkspaceRepo struct {
	ProviderID string
	Branch     string // branch for which applies are run

	*Webhook
}

// WorkspaceUpdateRepoOptions are options for updating a workspace's associated
// repo.
type WorkspaceUpdateRepoOptions struct {
	Branch *string
}

// unmarshal workspace repo
