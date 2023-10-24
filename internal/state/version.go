package state

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql/pggen"
)

const (
	Pending   Status = "pending"
	Finalized Status = "finalized"
	Discarded Status = "discarded"
)

var (
	ErrSerialNotGreaterThanCurrent = errors.New("the serial provided in the state file is not greater than the serial currently known remotely")
	ErrSerialMD5Mismatch           = errors.New("the MD5 hash of the state provided does not match what is currently known for the same serial number")
)

type (
	Status string

	// Version is a specific version of terraform state. It includes important
	// metadata as well as the state file itself.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions
	Version struct {
		ID          string             `jsonapi:"primary,state-versions"`
		CreatedAt   time.Time          `jsonapi:"attribute" json:"created-at"`
		Serial      int64              `jsonapi:"attribute" json:"serial"`
		State       []byte             `jsonapi:"attribute" json:"state"`
		Status      Status             `jsonapi:"attribute" json:"status"`
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
		State       []byte  // Terraform state file. Optional.
		WorkspaceID *string // ID of state version's workspace. Required.
		Serial      *int64  // State serial number. If not provided then it is extracted from the state.
	}

	// newVersionOptions are options for constructing a state version - options
	// are assumed to have already been validated.
	newVersionOptions struct {
		state       []byte
		workspaceID string
		serial      int64
	}

	// factory creates state versions - creation requires pre-requisite checking
	// with the db, hence necessity for a factory.
	factory struct {
		db factoryDB
	}

	factoryDB interface {
		Tx(context.Context, func(context.Context, pggen.Querier) error) error

		createVersion(context.Context, *Version) error
		getVersion(ctx context.Context, svID string) (*Version, error)
		getCurrentVersion(ctx context.Context, workspaceID string) (*Version, error)
		updateCurrentVersion(context.Context, string, string) error
	}
)

func (f *factory) new(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	if opts.WorkspaceID == nil {
		return nil, &internal.MissingParameterError{Parameter: "workspace_id"}
	}

	// NOTE: state file is optional
	// TODO: make the serial option mandatory
	// TODO: ensure serial option matches serial in file

	var (
		serial int64
	)
	// serial provided in options takes precedence over that extracted from the
	// state file.
	if opts.Serial != nil {
		serial = *opts.Serial
	} else if opts.State != nil {
		var file File
		if err := json.Unmarshal(opts.State, file); err != nil {
			return nil, err
		}
		serial = file.Serial
	} else {
		return nil, errors.New("either serial or state file must be provided")
	}

	// Serial should be greater than or equal to current serial
	current, err := f.db.getCurrentVersion(ctx, *opts.WorkspaceID)
	if errors.Is(err, internal.ErrResourceNotFound) {
		// this is the first state version for workspace, so set current serial
		// to a negative number to ensure tests below succeed.
		current = &Version{Serial: -1}
	} else if err != nil {
		return nil, err
	}
	if current.Serial > serial {
		return nil, ErrSerialNotGreaterThanCurrent
	}
	if current.Serial == serial {
		// Same serial is permissible as long as the state is identical. (This
		// follows the observed but undocumented behaviour of TFC).
		// If no state has been provided then an error is returned.
		if opts.State == nil {
			return nil, ErrSerialNotGreaterThanCurrent
		}
		if fmt.Sprintf("%x", md5.Sum(current.State)) != fmt.Sprintf("%x", md5.Sum(opts.State)) {
			return nil, ErrSerialMD5Mismatch
		}
	}

	if err := f.createCurrent(ctx, &sv); err != nil {
		return nil, err
	}
	return &sv, nil
}

// Create a state version and update workspace's current state version.
func (f *factory) createCurrent(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	sv := Version{
		ID:          internal.NewID("sv"),
		CreatedAt:   internal.CurrentTimestamp(),
		Serial:      opts.serial,
		State:       opts.state,
		WorkspaceID: opts.workspaceID,
	}

	err := f.db.Lock(ctx, "state_versions", func(ctx context.Context, q pggen.Querier) error {
		if err := f.db.createVersion(ctx, &sv); err != nil {
			return err
		}
		// optionally upload state and outputs
		if opts.State != nil {
			outputs, err := f.uploadStateAndOutputs(ctx, sv.ID, opts.State)
			if err != nil {
				return err
			}
			sv.Outputs = outputs
		}
		if err := f.db.updateCurrentVersion(ctx, sv.WorkspaceID, sv.ID); err != nil {
			return fmt.Errorf("updating current version: %w", err)
		}
		return nil
	})
	return &sv, err
}

func (f *factory) uploadStateAndOutputs(ctx context.Context, svID string, state []byte) ([]*Output, error) {
	// extract outputs from state file
	//
	// TODO: TFC performs this as an asynchronous task, maybe OTF should too...
	var file File
	if err := json.Unmarshal(opts.State, &file); err != nil {
		return nil, err
	}
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

	return f.db.Lock(ctx, "state_versions", func(ctx context.Context, q pggen.Querier) error {
		// TODO: check sv has status=pending
		if err := f.db.updateState(ctx, svID, state); err != nil {
			return err
		}
		// TODO: discard all svs with status=pending
		return nil
	})
}

func (f *factory) rollback(ctx context.Context, svID string) (*Version, error) {
	sv, err := f.db.getVersion(ctx, svID)
	if err != nil {
		return nil, err
	}
	return f.createCurrent(ctx, CreateStateVersionOptions{
		State:       v.State,
		WorkspaceID: v.WorkspaceID,
		Serial:      v.Serial,
	})
}

func (v *Version) String() string { return v.ID }

func (v *Version) File() (*File, error) {
	var f File
	if err := json.Unmarshal(v.State, &f); err != nil {
		return nil, err
	}
	return &f, nil
}
