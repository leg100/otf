package state

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
)

var (
	ErrSerialLessThanCurrent = errors.New("the serial provided in the state file is not greater than the serial currently known remotely")
	ErrSerialMD5Mismatch     = errors.New("the MD5 hash of the state provided does not match what is currently known for the same serial number")
)

type (
	// factory creates state versions - creation requires pre-requisite checking
	// with the db, hence necessity for a factory.
	factory struct {
		db db
	}

	// newVersionOptions are options for constructing a state version - options
	// are assumed to have already been validated.
	newVersionOptions struct {
		state       []byte
		workspaceID string
		serial      int64
	}
)

func (fa *factory) create(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	if opts.State == nil {
		return nil, &internal.MissingParameterError{Parameter: "state"}
	}
	if opts.WorkspaceID == nil {
		return nil, &internal.MissingParameterError{Parameter: "workspace_id"}
	}

	var f File
	if err := json.Unmarshal(opts.State, &f); err != nil {
		return nil, err
	}

	// Serial provided in options takes precedence over that extracted from the
	// state file.
	var serial int64
	if opts.Serial != nil {
		serial = *opts.Serial
	} else {
		serial = f.Serial
	}

	// Serial should be greater than or equal to current serial
	current, err := fa.db.getCurrentVersion(ctx, *opts.WorkspaceID)
	if errors.Is(err, internal.ErrResourceNotFound) {
		// this is the first state version for workspace, so set current serial
		// to a negative number to ensure tests below succeed.
		current = &Version{Serial: -1}
	} else if err != nil {
		return nil, err
	}
	if current.Serial > serial {
		return nil, ErrSerialLessThanCurrent
	}
	if current.Serial == serial {
		// Same serial is permissible as long as the state is identical. (This
		// follows the observed but undocumented behaviour of TFC).
		if fmt.Sprintf("%x", md5.Sum(current.State)) != fmt.Sprintf("%x", md5.Sum(opts.State)) {
			return nil, ErrSerialMD5Mismatch
		}
	}

	sv, err := newVersion(newVersionOptions{
		state:       opts.State,
		workspaceID: *opts.WorkspaceID,
		serial:      serial,
	})
	if err != nil {
		return nil, err
	}

	if err := fa.createCurrent(ctx, &sv); err != nil {
		return nil, err
	}
	return &sv, nil
}

// Create a state version and update workspace's current state version.
func (fa *factory) createCurrent(ctx context.Context, sv *Version) error {
	return fa.db.tx(ctx, func(tx db) error {
		if err := tx.createVersion(ctx, sv); err != nil {
			return err
		}
		if err := tx.updateCurrentVersion(ctx, sv.WorkspaceID, sv.ID); err != nil {
			return fmt.Errorf("updating current version: %w", err)
		}
		return nil
	})
}

func (fa *factory) rollback(ctx context.Context, svID string) (*Version, error) {
	sv, err := fa.db.getVersion(ctx, svID)
	if err != nil {
		return nil, err
	}
	clone, err := sv.Clone()
	if err != nil {
		return nil, err
	}
	if err := fa.createCurrent(ctx, clone); err != nil {
		return nil, err
	}
	return clone, nil
}

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
	outputs := make(OutputList, len(f.Outputs))
	for k, v := range f.Outputs {
		hclType, err := newHCLType(v.Value)
		if err != nil {
			return Version{}, err
		}

		outputs[k] = &Output{
			ID:             internal.NewID("wsout"),
			Name:           k,
			Type:           hclType,
			Value:          v.Value,
			Sensitive:      v.Sensitive,
			StateVersionID: sv.ID,
		}
	}
	sv.Outputs = outputs

	return sv, nil
}
