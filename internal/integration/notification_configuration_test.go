package integration

import (
	"errors"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/notifications"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_NotificationConfigurationService(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		_, err := svc.CreateNotificationConfiguration(ctx, ws.ID, notifications.CreateConfigOptions{
			DestinationType: notifications.DestinationGeneric,
			Enabled:         internal.Bool(true),
			Name:            internal.String("testing"),
		})
		require.NoError(t, err)
	})

	t.Run("update", func(t *testing.T) {
		svc := setup(t, nil)
		nc := svc.createNotificationConfig(t, ctx, nil)

		got, err := svc.UpdateNotificationConfiguration(ctx, nc.ID, notifications.UpdateConfigOptions{
			Name:    internal.String("new-name"),
			Enabled: internal.Bool(false),
		})
		require.NoError(t, err)

		assert.Equal(t, "new-name", got.Name)
		assert.False(t, got.Enabled)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		nc1 := svc.createNotificationConfig(t, ctx, ws)
		nc2 := svc.createNotificationConfig(t, ctx, ws)
		nc3 := svc.createNotificationConfig(t, ctx, ws)

		got, err := svc.ListNotificationConfigurations(ctx, ws.ID)
		require.NoError(t, err)

		assert.Equal(t, 3, len(got))
		assert.Contains(t, got, nc1)
		assert.Contains(t, got, nc2)
		assert.Contains(t, got, nc3)
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)
		nc := svc.createNotificationConfig(t, ctx, nil)

		got, err := svc.GetNotificationConfiguration(ctx, nc.ID)
		require.NoError(t, err)

		assert.Equal(t, nc, got)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		nc := svc.createNotificationConfig(t, ctx, nil)

		err := svc.DeleteNotificationConfiguration(ctx, nc.ID)
		require.NoError(t, err)

		_, err = svc.GetNotificationConfiguration(ctx, nc.ID)
		require.True(t, errors.Is(err, internal.ErrResourceNotFound))
	})
}
