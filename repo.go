package otf

import (
	"context"

	"github.com/google/uuid"
)

const (
	WorkspaceConnection ConnectionType = iota
	ModuleConnection
)

type (
	// ConnectionType identifies the OTF resource type in a VCS connection.
	ConnectionType int

	// Connection is a connection between a VCS repo and an OTF resource.
	Connection struct {
		VCSProviderID string
		WebhookID     uuid.UUID
		// TODO: this is redundant info (identifier can be retrieved via
		// webhookID)
		Identifier string // identifier is <repo_owner>/<repo_name>
	}

	// RepoService manages configuration of VCS repositories
	RepoService interface {
		// Connect adds a connection between a VCS repo and an OTF resource. A
		// webhook is created if one doesn't exist already.
		Connect(ctx context.Context, opts ConnectOptions) (*Connection, error)
		// Disconnect removes a connection between a VCS repo and an OTF
		// resource. If there are no more connections then its
		// webhook is removed.
		Disconnect(ctx context.Context, opts DisconnectOptions) error
	}

	ConnectOptions struct {
		ConnectionType // OTF resource type

		VCSProviderID string // vcs provider of repo
		ResourceID    string // ID of OTF resource
		Identifier    string // Repo path
		Tx            DB     // Optional tx for performing database ops within.
	}

	DisconnectOptions struct {
		ConnectionType // OTF resource type

		ResourceID string // ID of OTF resource
		Tx         DB     // Optional tx for performing database ops within.
	}
)
