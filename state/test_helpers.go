package state

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

type fakeDB struct {
	current *Version // returned by getCurrentVersion
	version *Version // returned by getVersion
	db
}

func (f *fakeDB) getVersion(ctx context.Context, svID string) (*Version, error) {
	if f.version == nil {
		return nil, otf.ErrResourceNotFound
	}
	return f.version, nil
}

func (f *fakeDB) getCurrentVersion(ctx context.Context, workspaceID string) (*Version, error) {
	if f.current == nil {
		return nil, otf.ErrResourceNotFound
	}
	return f.current, nil
}

func (f *fakeDB) createVersion(ctx context.Context, v *Version) error {
	return nil
}

func (f *fakeDB) updateCurrentVersion(ctx context.Context, workspaceID, svID string) error {
	return nil
}

func (f *fakeDB) tx(ctx context.Context, txfunc func(db) error) error {
	return txfunc(f)
}

func compactJSON(t *testing.T, src string) string {
	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(src))
	require.NoError(t, err)
	return buf.String()
}
