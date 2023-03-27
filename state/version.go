package state

import (
	"encoding/json"
	"errors"
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

// newVersion constructs a new state version.
func newVersion(opts CreateStateVersionOptions) (*Version, error) {
	if opts.State == nil {
		return nil, errors.New("state file required")
	}
	if opts.WorkspaceID == nil {
		return nil, errors.New("workspace ID required")
	}

	var f file
	if err := json.Unmarshal(opts.State, &f); err != nil {
		return nil, err
	}

	sv := Version{
		ID:          otf.NewID("sv"),
		CreatedAt:   otf.CurrentTimestamp(),
		Serial:      f.Serial,
		State:       opts.State,
		WorkspaceID: *opts.WorkspaceID,
	}
	// Serial provided in options takes precedence over that extracted from the
	// state file.
	if opts.Serial != nil {
		sv.Serial = *opts.Serial
	}

	sv.Outputs = make(outputList, len(f.Outputs))
	for k, v := range f.Outputs {
		hclType, err := newHCLType(v.Value)
		if err != nil {
			return nil, err
		}

		sv.Outputs[k] = &Output{
			ID:             otf.NewID("wsout"),
			Name:           k,
			Type:           hclType,
			Value:          string(v.Value),
			Sensitive:      v.Sensitive,
			StateVersionID: sv.ID,
		}
	}
	return &sv, nil
}

func (v *Version) String() string { return v.ID }
