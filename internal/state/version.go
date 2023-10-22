package state

import (
	"encoding/json"
	"time"

	"github.com/leg100/otf/internal"
)

type (
	// Version is a specific Version of terraform state. It includes important
	// metadata as well as the state file itself.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions
	Version struct {
		ID          string             `jsonapi:"primary,state-versions"`
		CreatedAt   time.Time          `jsonapi:"attribute" json:"created-at"`
		Serial      int64              `jsonapi:"attribute" json:"serial"`
		State       []byte             `jsonapi:"attribute" json:"state"`
		Outputs     map[string]*Output `jsonapi:"attribute" json:"outputs"`
		WorkspaceID string             `jsonapi:"attribute" json:"workspace-id"`
	}

	Output struct {
		ID             string
		Name           string
		Type           string
		Value          json.RawMessage
		Sensitive      bool
		StateVersionID string
	}

	// CreateStateVersionOptions are options for creating a state version.
	CreateStateVersionOptions struct {
		State       []byte  // Terraform state file. Required.
		WorkspaceID *string // ID of state version's workspace. Required.
		Serial      *int64  // State serial number. If not provided then it is extracted from the state.
	}
)

func newVersion(opts newVersionOptions) (Version, error) {
	sv := Version{
		ID:          internal.NewID("sv"),
		CreatedAt:   internal.CurrentTimestamp(),
		Serial:      opts.serial,
		State:       opts.state,
		WorkspaceID: opts.workspaceID,
	}

	var f File
	if err := json.Unmarshal(opts.state, &f); err != nil {
		return Version{}, err
	}

	// extract outputs from state file
	outputs := make(map[string]*Output, len(f.Outputs))
	for k, v := range f.Outputs {
		typ, err := v.Type()
		if err != nil {
			return Version{}, err
		}

		outputs[k] = &Output{
			ID:             internal.NewID("wsout"),
			Name:           k,
			Type:           typ,
			Value:          v.Value,
			Sensitive:      v.Sensitive,
			StateVersionID: sv.ID,
		}
	}
	sv.Outputs = outputs

	return sv, nil
}

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

func (v *Version) File() (*File, error) {
	var f File
	if err := json.Unmarshal(v.State, &f); err != nil {
		return nil, err
	}
	return &f, nil
}
