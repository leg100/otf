package state

import (
	"time"

	internal "github.com/leg100/otf"
)

type (
	// Version is a specific Version of terraform state. It includes important
	// metadata as well as the state file itself.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions
	Version struct {
		ID          string
		CreatedAt   time.Time
		Serial      int64
		State       []byte     // state file
		Outputs     OutputList // state version has many outputs
		WorkspaceID string     // state version belongs to a workspace
	}

	// VersionList represents a list of state versions.
	VersionList struct {
		*internal.Pagination
		Items []*Version
	}

	Output struct {
		ID             string
		Name           string
		Type           string
		Value          string
		Sensitive      bool
		StateVersionID string
	}

	OutputList map[string]*Output

	// CreateStateVersionOptions are options for creating a state version.
	CreateStateVersionOptions struct {
		State       []byte  // Terraform state file. Required.
		WorkspaceID *string // ID of state version's workspace. Required.
		Serial      *int64  // State serial number. If not provided then it is extracted from the state.
	}
)

func (v *Version) String() string { return v.ID }

// Clone makes a copy albeit with new identifiers.
func (v *Version) Clone() (*Version, error) {
	cloned, err := newVersion(newVersionOptions{
		state:       v.State,
		workspaceID: v.WorkspaceID,
		serial:      v.Serial,
	})
	if err != nil {
		return nil, err
	}
	return &cloned, nil
}
