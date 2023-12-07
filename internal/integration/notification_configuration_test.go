package integration

import (
	"errors"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_NotificationConfigurationService(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		daemon, org, ctx := setup(t, nil)
		ws := daemon.createWorkspace(t, ctx, org)
		sub, unsub := daemon.WatchNotificationConfigurations(ctx)
		defer unsub()
		nc, err := daemon.CreateNotificationConfiguration(ctx, ws.ID, notifications.CreateConfigOptions{
			DestinationType: notifications.DestinationGeneric,
			Enabled:         internal.Bool(true),
			Name:            internal.String("testing"),
			URL:             internal.String("http://example.com"),
		})
		require.NoError(t, err)
		assert.Equal(t, pubsub.NewCreatedEvent(nc), <-sub)
	})

	t.Run("update", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		nc := svc.createNotificationConfig(t, ctx, nil)

		t.Run("name", func(t *testing.T) {
			got, err := svc.UpdateNotificationConfiguration(ctx, nc.ID, notifications.UpdateConfigOptions{
				Name: internal.String("new-name"),
			})
			require.NoError(t, err)
			assert.Equal(t, "new-name", got.Name)
		})

		t.Run("disable", func(t *testing.T) {
			got, err := svc.UpdateNotificationConfiguration(ctx, nc.ID, notifications.UpdateConfigOptions{
				Enabled: internal.Bool(false),
			})
			require.NoError(t, err)
			assert.False(t, got.Enabled)
		})

		t.Run("url", func(t *testing.T) {
			got, err := svc.UpdateNotificationConfiguration(ctx, nc.ID, notifications.UpdateConfigOptions{
				URL: internal.String("http://otf.ninja/notifications"),
			})
			require.NoError(t, err)
			assert.Equal(t, internal.String("http://otf.ninja/notifications"), got.URL)
		})
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
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
		svc, _, ctx := setup(t, nil)
		nc := svc.createNotificationConfig(t, ctx, nil)

		got, err := svc.GetNotificationConfiguration(ctx, nc.ID)
		require.NoError(t, err)

		assert.Equal(t, nc, got)
	})

	t.Run("delete", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, org)
		sub, unsub := svc.WatchNotificationConfigurations(ctx)
		defer unsub()
		nc := svc.createNotificationConfig(t, ctx, ws)
		assert.Equal(t, pubsub.NewCreatedEvent(nc), <-sub)

		err := svc.DeleteNotificationConfiguration(ctx, nc.ID)
		require.NoError(t, err)
		assert.Equal(t, pubsub.NewDeletedEvent(&notifications.Config{ID: nc.ID}), <-sub)

		_, err = svc.GetNotificationConfiguration(ctx, nc.ID)
		require.True(t, errors.Is(err, internal.ErrResourceNotFound))
	})

	// test the postgres' ON DELETE CASCADE functionality as well as postgres
	// event triggers: when a workspace is deleted, its notification
	// configurations should be deleted too and events should be sent out.
	t.Run("cascade delete", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)
		sub, unsub := svc.WatchNotificationConfigurations(ctx)
		defer unsub()

		ws := svc.createWorkspace(t, ctx, org)

		nc1 := svc.createNotificationConfig(t, ctx, ws)
		assert.Equal(t, pubsub.NewCreatedEvent(nc1), <-sub)

		nc2 := svc.createNotificationConfig(t, ctx, ws)
		assert.Equal(t, pubsub.NewCreatedEvent(nc2), <-sub)

		_, err := svc.Workspaces.DeleteWorkspace(ctx, ws.ID)
		require.NoError(t, err)

		assert.Equal(t, pubsub.NewDeletedEvent(&notifications.Config{ID: nc1.ID}), <-sub)
		assert.Equal(t, pubsub.NewDeletedEvent(&notifications.Config{ID: nc2.ID}), <-sub)
	})
}
