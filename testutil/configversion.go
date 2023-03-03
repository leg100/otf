package testutil

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	"github.com/stretchr/testify/require"
)

func NewTestConfigurationVersion(t *testing.T, ws otf.Workspace, opts otf.ConfigurationVersionCreateOptions) *ConfigurationVersion {
	cv, err := NewConfigurationVersion(ws.ID, opts)
	require.NoError(t, err)
	return cv
}

func CreateConfigurationVersion(t *testing.T, db otf.DB, ws otf.Workspace, opts otf.ConfigurationVersionCreateOptions) otf.ConfigurationVersion {
	ctx := context.Background()
	configService := NewConfigVersionService(db)

	cv, err := configService.CreateConfigurationVersion(ctx, ws.ID, otf.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)

	t.Cleanup(func() {
		configService.DeleteConfigurationVersion(ctx, cv.ID)
	})
	return cv
}

func NewConfigVersionService(db otf.DB) *configversion.Service {
	return configversion.NewService(configversion.Options{
		Authorizer: NewAllowAllAuthorizer(),
		Cache:      newFakeCache(),
		Database:   db,
		Logger:     logr.Discard(),
	})
}
