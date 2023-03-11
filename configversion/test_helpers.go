package configversion

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

func NewTestConfigurationVersion(t *testing.T, ws *workspace.Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
	cv, err := NewConfigurationVersion(ws.ID, opts)
	require.NoError(t, err)
	return cv
}

func CreateTestConfigurationVersion(t *testing.T, db otf.DB, ws *workspace.Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
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

type fakeService struct {
	Service
}

func (f *fakeService) upload(context.Context, string, []byte) error {
	return nil
}
