package otf

import (
	"context"
)

const (
	WorkspaceConnection ConnectionType = iota
	ModuleConnection
)

type (
	// ConnectionType identifies the OTF resource type in a VCS connection.
	ConnectionType int

	// RepoService manages configuration of VCS repositories
	RepoService interface {
		// Connect adds a connection between a VCS repo and an OTF resource. A
		// webhook is created if one doesn't exist already.
		Connect(ctx context.Context, opts ConnectionOptions) error
		// Disconnect removes a connection between a VCS repo and an OTF
		// resource. If there are no more connections then its
		// webhook is removed.
		Disconnect(ctx context.Context, opts ConnectionOptions) error
	}

	ConnectionOptions struct {
		ConnectionType // OTF resource type

		VCSProviderID string // vcs provider of repo
		ResourceID    string // ID of OTF resource
		Identifier    string // Repo path
		Cloud         string // Cloud hosting the repo
		Tx            DB     // database tx for adding/removing connection in OTF db
	}
)
