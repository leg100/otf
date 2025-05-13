package state

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"golang.org/x/exp/maps"
)

const (
	Pending   Status = "pending"
	Finalized Status = "finalized"
	Discarded Status = "discarded"
)

var (
	ErrSerialNotGreaterThanCurrent = errors.New("the serial provided in the state file is not greater than the serial currently known remotely")
	ErrSerialMD5Mismatch           = errors.New("the MD5 hash of the state provided does not match what is currently known for the same serial number")
	ErrUploadNonPending            = errors.New("cannot upload state to a state version with a non-pending status")
)

type (
	Status string

	// Version is a specific version of terraform state. It includes important
	// metadata as well as the state file itself.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions
	Version struct {
		ID          resource.TfeID     `jsonapi:"primary,state-versions"`
		CreatedAt   time.Time          `jsonapi:"attribute" json:"created-at"`
		Serial      int64              `jsonapi:"attribute" json:"serial"`
		State       []byte             `jsonapi:"attribute" json:"state"`
		Status      Status             `jsonapi:"attribute" json:"status"`
		Outputs     map[string]*Output `jsonapi:"attribute" json:"outputs"`
		WorkspaceID resource.TfeID     `jsonapi:"attribute" json:"workspace-id"`
	}

	Output struct {
		ID             resource.TfeID
		Name           string
		Type           string
		Value          json.RawMessage
		Sensitive      bool
		StateVersionID resource.TfeID
	}

	// CreateStateVersionOptions are options for creating a state version.
	CreateStateVersionOptions struct {
		State       []byte         // Terraform state file. Optional.
		WorkspaceID resource.TfeID // ID of state version's workspace. Required.
		Serial      *int64         // State serial number. Required.
	}

	// factory creates state versions - creation requires pre-requisite checking
	// with the db, hence necessity for a factory.
	factory struct {
		db factoryDB
	}

	factoryDB interface {
		Tx(context.Context, func(context.Context) error) error

		createVersion(context.Context, *Version) error
		createOutputs(context.Context, []*Output) error
		getVersion(ctx context.Context, svID resource.ID) (*Version, error)
		getCurrentVersion(ctx context.Context, workspaceID resource.TfeID) (*Version, error)
		updateCurrentVersion(context.Context, resource.TfeID, resource.TfeID) error
		uploadStateAndFinalize(ctx context.Context, svID resource.TfeID, state []byte) error
		discardAnyPending(ctx context.Context, workspaceID resource.TfeID) error
	}
)

// new create a new state version
func (f *factory) new(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	if opts.Serial == nil {
		return nil, &internal.ErrMissingParameter{Parameter: "serial"}
	}
	// Serial should be greater than or equal to current serial
	current, err := f.db.getCurrentVersion(ctx, opts.WorkspaceID)
	if errors.Is(err, internal.ErrResourceNotFound) {
		// this is the first state version for workspace, so set current serial
		// to a negative number to ensure tests below succeed.
		current = &Version{Serial: -1}
	} else if err != nil {
		return nil, fmt.Errorf("retrieving current version: %w", err)
	}
	if current.Serial > *opts.Serial {
		return nil, ErrSerialNotGreaterThanCurrent
	}
	if current.Serial == *opts.Serial {
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
	return f.newWithoutValidation(ctx, opts)
}

// newWithoutValidation creates a state version without validating the options.
func (f *factory) newWithoutValidation(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	sv := Version{
		ID:          resource.NewTfeID(resource.StateVersionKind),
		CreatedAt:   internal.CurrentTimestamp(nil),
		Serial:      *opts.Serial,
		State:       opts.State,
		Status:      Pending,
		WorkspaceID: opts.WorkspaceID,
	}
	err := f.db.Tx(ctx, func(ctx context.Context) error {
		if err := f.db.createVersion(ctx, &sv); err != nil {
			return fmt.Errorf("creating version in database: %w", err)
		}
		if opts.State != nil {
			finalized, err := f.uploadStateAndOutputs(ctx, &sv, opts.State)
			if err != nil {
				return fmt.Errorf("uploading state to database: %w", err)
			}
			sv = *finalized
		}
		return nil
	})
	return &sv, err
}

// upload state and its outputs to the database
func (f *factory) uploadStateAndOutputs(ctx context.Context, sv *Version, state []byte) (*Version, error) {
	// extract outputs from state file
	//
	// TODO: TFC performs this as an asynchronous task, maybe OTF should too.
	var file File
	if err := json.Unmarshal(state, &file); err != nil {
		return nil, err
	}
	outputs := make(map[string]*Output, len(file.Outputs))
	for k, v := range file.Outputs {
		typ, err := v.Type()
		if err != nil {
			return nil, err
		}
		outputs[k] = &Output{
			ID:             resource.NewTfeID(resource.StateVersionOutputKind),
			Name:           k,
			Type:           typ,
			Value:          v.Value,
			Sensitive:      v.Sensitive,
			StateVersionID: sv.ID,
		}
	}
	// now perform database updates
	err := f.db.Tx(ctx, func(ctx context.Context) (err error) {
		if sv.Status != Pending {
			return ErrUploadNonPending
		}
		if err := f.db.createOutputs(ctx, maps.Values(outputs)); err != nil {
			return fmt.Errorf("creating outputs: %w", err)
		}
		if err := f.db.uploadStateAndFinalize(ctx, sv.ID, state); err != nil {
			return fmt.Errorf("uploading state: %w", err)
		}
		if err := f.db.discardAnyPending(ctx, sv.WorkspaceID); err != nil {
			return fmt.Errorf("discarding pending versions: %w", err)
		}
		if err := f.db.updateCurrentVersion(ctx, sv.WorkspaceID, sv.ID); err != nil {
			return fmt.Errorf("updating current version: %w", err)
		}
		return nil
	})
	// ensure state version reflects changes made via database.
	sv.Status = Finalized
	sv.Outputs = outputs
	return sv, err
}

func (f *factory) rollback(ctx context.Context, svID resource.TfeID) (*Version, error) {
	sv, err := f.db.getVersion(ctx, svID)
	if err != nil {
		return nil, err
	}
	return f.newWithoutValidation(ctx, CreateStateVersionOptions{
		State:       sv.State,
		WorkspaceID: sv.WorkspaceID,
		Serial:      &sv.Serial,
	})
}

func (v *Version) String() string { return v.ID.String() }

func (v *Version) File() (*File, error) {
	var f File
	if err := json.Unmarshal(v.State, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (v *Version) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ID.String()),
		slog.Int64("serial", v.Serial),
		slog.String("status", string(v.Status)),
		slog.String("workspace_id", v.WorkspaceID.String()),
	}
	return slog.GroupValue(attrs...)
}
