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

func NewTestConfigurationVersion(t *testing.T, ws otf.Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
	cv, err := NewConfigurationVersion(ws.ID(), opts)
	require.NoError(t, err)
	return cv
}

func CreateTestConfigurationVersion(t *testing.T, db otf.DB, ws otf.Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
	ctx := context.Background()
	cv := NewTestConfigurationVersion(t, ws, opts)
	err := db.CreateConfigurationVersion(ctx, cv)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteConfigurationVersion(ctx, cv.ID())
	})
	return cv
}
