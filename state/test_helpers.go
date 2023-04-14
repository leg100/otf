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
	current *Version
	db
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
