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

func NewTestConfigurationVersion(t *testing.T, ws *Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
	cv, err := NewConfigurationVersion(ws.ID(), opts)
	require.NoError(t, err)
	return cv
}
