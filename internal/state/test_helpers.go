package state

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type fakeDB struct {
	current *Version // returned by getCurrentVersion
	version *Version // returned by getVersion
}

func (f *fakeDB) Tx(ctx context.Context, fn func(context.Context, sql.Connection) error) error {
	return fn(ctx, nil)
}

func (f *fakeDB) createVersion(context.Context, *Version) error {
	return nil
}

func (f *fakeDB) createOutputs(context.Context, []*Output) error {
	return nil
}

func (f *fakeDB) getVersion(ctx context.Context, svID resource.TfeID) (*Version, error) {
	if f.version == nil {
		return nil, internal.ErrResourceNotFound
	}
	return f.version, nil
}

func (f *fakeDB) getCurrentVersion(ctx context.Context, workspaceID resource.TfeID) (*Version, error) {
	if f.current == nil {
		return nil, internal.ErrResourceNotFound
	}
	return f.current, nil
}

func (f *fakeDB) discardAnyPending(ctx context.Context, workspaceID resource.TfeID) error {
	return nil
}

func (f *fakeDB) updateCurrentVersion(ctx context.Context, workspaceID, svID resource.TfeID) error {
	return nil
}

func (f *fakeDB) uploadStateAndFinalize(ctx context.Context, svID resource.TfeID, state []byte) error {
	return nil
}
