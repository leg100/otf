package state

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leg100/otf"
)

var (
	ErrSerialLessThanCurrent = errors.New("the serial provided in the state file is not greater than the serial currently known remotely")
	ErrSerialMD5Mismatch     = errors.New("the MD5 hash of the state provided does not match what is currently known for the same serial number")
)

type (
	factory struct {
		db db
	}
)

func (fa *factory) create(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	if opts.State == nil {
		return nil, &otf.MissingParameterError{Parameter: "state"}
	}
	if opts.WorkspaceID == nil {
		return nil, &otf.MissingParameterError{Parameter: "workspace_id"}
	}

	var f file
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
	if errors.Is(err, otf.ErrResourceNotFound) {
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

	sv := Version{
		ID:          otf.NewID("sv"),
		CreatedAt:   otf.CurrentTimestamp(),
		Serial:      serial,
		State:       opts.State,
		WorkspaceID: *opts.WorkspaceID,
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

	// Create a state version and update workspace's current state version.
	err = fa.db.tx(ctx, func(tx db) error {
		if err := tx.createVersion(ctx, &sv); err != nil {
			return err
		}
		if err := tx.updateCurrentVersion(ctx, *opts.WorkspaceID, sv.ID); err != nil {
			return fmt.Errorf("updating current version: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &sv, nil
}
