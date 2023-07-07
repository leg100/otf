package state

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql/pggen"
)

type fakeDB struct {
	current *Version // returned by getCurrentVersion
	version *Version // returned by getVersion
	factoryDB
}

func (f *fakeDB) getVersion(ctx context.Context, svID string) (*Version, error) {
	if f.version == nil {
		return nil, internal.ErrResourceNotFound
	}
	return f.version, nil
}

func (f *fakeDB) getCurrentVersion(ctx context.Context, workspaceID string) (*Version, error) {
	if f.current == nil {
		return nil, internal.ErrResourceNotFound
	}
	return f.current, nil
}

func (f *fakeDB) Tx(context.Context, func(context.Context, pggen.Querier) error) error {
	return nil
}
