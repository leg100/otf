package state

import (
	"time"

	"github.com/leg100/otf"
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
		Outputs     outputList // state version has many outputs
		WorkspaceID string     // state version belongs to a workspace
	}

	// VersionList represents a list of state versions.
	VersionList struct {
		*otf.Pagination
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

	outputList map[string]*Output

	CreateStateVersionOptions struct {
		State       []byte  // Terraform state file. Required.
		WorkspaceID *string // ID of state version's workspace. Required.
		Serial      *int64  // State serial number. If not provided then it is extracted from the state.
	}
)

func (v *Version) String() string { return v.ID }
