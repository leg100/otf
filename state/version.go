package state

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/leg100/otf"
)

type (
	// version is a specific version of terraform state. It includes important
	// metadata as well as the state file itself.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions
	version struct {
		ID          string
		CreatedAt   time.Time
		Serial      int64
		State       []byte     // state file
		Outputs     outputList // state version has many outputs
		WorkspaceID string     // state version belongs to a workspace
	}

	// versionList represents a list of state versions.
	versionList struct {
		*otf.Pagination
		Items []*version
	}

	output struct {
		id             string
		name           string
		typ            string
		value          string
		sensitive      bool
		stateVersionID string
	}

	outputList map[string]*output
)

// newVersion constructs a new state version.
func newVersion(opts CreateStateVersionOptions) (*version, error) {
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

	sv := version{
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

		sv.Outputs[k] = &output{
			id:             otf.NewID("wsout"),
			name:           k,
			typ:            hclType,
			value:          string(v.Value),
			sensitive:      v.Sensitive,
			stateVersionID: sv.ID,
		}
	}
	return &sv, nil
}

func (v *version) String() string { return v.ID }
