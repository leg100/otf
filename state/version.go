package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf"
)

// version is a specific version of terraform state. It includes important
// metadata as well as the state file itself.
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions
type version struct {
	id          string
	createdAt   time.Time
	serial      int64
	state       []byte     // state file
	outputs     outputList // state version has many outputs
	workspaceID string     // state version belongs to a workspace
}

// newVersion constructs a new state version.
func newVersion(opts otf.CreateStateVersionOptions) (*version, error) {
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
		id:          otf.NewID("sv"),
		createdAt:   otf.CurrentTimestamp(),
		serial:      f.Serial,
		state:       opts.State,
		workspaceID: *opts.WorkspaceID,
	}
	// Serial provided in options takes precedence over that extracted from the
	// state file.
	if opts.Serial != nil {
		sv.serial = *opts.Serial
	}

	for k, v := range f.Outputs {
		hclType, err := newHCLType(v.Value)
		if err != nil {
			return nil, err
		}

		sv.outputs = append(sv.outputs, &output{
			id:             otf.NewID("wsout"),
			name:           k,
			typ:            hclType,
			value:          v.Value,
			sensitive:      v.Sensitive,
			stateVersionID: sv.id,
		})
	}
	return &sv, nil
}

func (v *version) ID() string           { return v.id }
func (v *version) CreatedAt() time.Time { return v.createdAt }
func (v *version) String() string       { return v.id }
func (v *version) Serial() int64        { return v.serial }
func (v *version) State() []byte        { return v.state }

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (v *version) ToJSONAPI() any {
	j := &jsonapiVersion{
		ID:          v.ID(),
		CreatedAt:   v.CreatedAt(),
		DownloadURL: fmt.Sprintf("/api/v2/state-versions/%s/download", v.ID()),
		Serial:      v.Serial(),
	}
	for _, out := range v.outputs {
		j.Outputs = append(j.Outputs, out.ToJSONAPI().(*jsonapiVersionOutput))
	}
	return j
}
