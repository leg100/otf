package configversion

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

type fakeConfigurationVersionApp struct {
	otf.Application
}

func (f *fakeConfigurationVersionApp) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func NewTestConfigurationVersion(t *testing.T, ws otf.Workspace, opts otf.ConfigurationVersionCreateOptions) *otf.ConfigurationVersion {
	cv, err := otf.NewConfigurationVersion(ws.ID, opts)
	require.NoError(t, err)
	return cv
}

func CreateTestConfigurationVersion(t *testing.T, db otf.DB, ws otf.Workspace, opts otf.ConfigurationVersionCreateOptions) *otf.ConfigurationVersion {
	ctx := context.Background()
	cv := NewTestConfigurationVersion(t, ws, opts)
	cvDB := newPGDB(db)

	err := cvDB.CreateConfigurationVersion(ctx, cv)
	require.NoError(t, err)

	t.Cleanup(func() {
		cvDB.DeleteConfigurationVersion(ctx, cv.ID)
	})
	return cv
}
