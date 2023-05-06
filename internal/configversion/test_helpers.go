package configversion

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

func NewTestConfigurationVersion(t *testing.T, ws *workspace.Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
	cv, err := NewConfigurationVersion(ws.ID, opts)
	require.NoError(t, err)
	return cv
}

func CreateTestConfigurationVersion(t *testing.T, db internal.DB, ws *workspace.Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
	ctx := context.Background()
	cv := NewTestConfigurationVersion(t, ws, opts)
	cvDB := &pgdb{db}

	err := cvDB.CreateConfigurationVersion(ctx, cv)
	require.NoError(t, err)

	t.Cleanup(func() {
		cvDB.DeleteConfigurationVersion(ctx, cv.ID)
	})
	return cv
}
