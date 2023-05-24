// Package repo handles configuration of VCS repositories.
package repo

import "github.com/leg100/otf/internal"

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
		Repo          string
	}

	ConnectOptions struct {
		ConnectionType // OTF resource type

		VCSProviderID string // vcs provider of repo
		ResourceID    string // ID of OTF resource
		RepoPath      string
		Tx            internal.DB // Optional tx for performing database ops within.
	}

	DisconnectOptions struct {
		ConnectionType // OTF resource type

		ResourceID string      // ID of OTF resource
		Tx         internal.DB // Optional tx for performing database ops within.
	}

	SynchroniseOptions struct {
		VCSProviderID string // vcs provider of repo
		RepoPath      string
	}
)
