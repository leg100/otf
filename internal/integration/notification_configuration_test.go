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
		daemon, org, ctx := setup(t)
		ws := daemon.createWorkspace(t, ctx, org)
		sub, unsub := daemon.Notifications.Watch(ctx)
		defer unsub()
		nc, err := daemon.Notifications.Create(ctx, ws.ID, notifications.CreateConfigOptions{
			DestinationType: notifications.DestinationGeneric,
			Enabled:         new(true),
			Name:            new("testing"),
			URL:             new("http://example.com"),
		})
		require.NoError(t, err)
		event := <-sub
		assert.Equal(t, pubsub.CreatedEvent, event.Type)
		assert.Equal(t, nc.ID, event.Payload.ID)
	})

	t.Run("update", func(t *testing.T) {
		svc, _, ctx := setup(t)
		nc := svc.createNotificationConfig(t, ctx, nil)

		t.Run("name", func(t *testing.T) {
			got, err := svc.Notifications.Update(ctx, nc.ID, notifications.UpdateConfigOptions{
				Name: new("new-name"),
			})
			require.NoError(t, err)
			assert.Equal(t, "new-name", got.Name)
		})

		t.Run("disable", func(t *testing.T) {
			got, err := svc.Notifications.Update(ctx, nc.ID, notifications.UpdateConfigOptions{
				Enabled: new(false),
			})
			require.NoError(t, err)
			assert.False(t, got.Enabled)
		})

		t.Run("url", func(t *testing.T) {
			got, err := svc.Notifications.Update(ctx, nc.ID, notifications.UpdateConfigOptions{
				URL: new("http://otf.ninja/notifications"),
			})
			require.NoError(t, err)
			assert.Equal(t, new("http://otf.ninja/notifications"), got.URL)
		})
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, nil)
		nc1 := svc.createNotificationConfig(t, ctx, ws)
		nc2 := svc.createNotificationConfig(t, ctx, ws)
		nc3 := svc.createNotificationConfig(t, ctx, ws)

		got, err := svc.Notifications.List(ctx, ws.ID)
		require.NoError(t, err)

		assert.Equal(t, 3, len(got))
		assert.Contains(t, got, nc1)
		assert.Contains(t, got, nc2)
		assert.Contains(t, got, nc3)
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t)
		nc := svc.createNotificationConfig(t, ctx, nil)

		got, err := svc.Notifications.Get(ctx, nc.ID)
		require.NoError(t, err)

		assert.Equal(t, nc, got)
	})

	t.Run("delete", func(t *testing.T) {
		svc, org, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, org)
		sub, unsub := svc.Notifications.Watch(ctx)
		defer unsub()
		nc := svc.createNotificationConfig(t, ctx, ws)

		// dismiss created event
		<-sub

		err := svc.Notifications.Delete(ctx, nc.ID)
		require.NoError(t, err)

		// should receive deleted event
		event := <-sub
		assert.Equal(t, pubsub.DeletedEvent, event.Type)
		assert.Equal(t, nc.ID, event.Payload.ID)

		_, err = svc.Notifications.Get(ctx, nc.ID)
		require.True(t, errors.Is(err, internal.ErrResourceNotFound))
	})

	// test the postgres' ON DELETE CASCADE functionality as well as postgres
	// event triggers: when a workspace is deleted, its notification
	// configurations should be deleted too and events should be sent out.
	t.Run("cascade delete", func(t *testing.T) {
		svc, org, ctx := setup(t)
		sub, unsub := svc.Notifications.Watch(ctx)
		defer unsub()

		ws := svc.createWorkspace(t, ctx, org)

		nc1 := svc.createNotificationConfig(t, ctx, ws)
		// dismiss created event
		<-sub

		nc2 := svc.createNotificationConfig(t, ctx, ws)
		// dismiss created event
		<-sub

		_, err := svc.Workspaces.Delete(ctx, ws.ID)
		require.NoError(t, err)

		// should receive deleted events
		event := <-sub
		assert.Equal(t, pubsub.DeletedEvent, event.Type)
		assert.Equal(t, nc1.ID, event.Payload.ID)

		event = <-sub
		assert.Equal(t, pubsub.DeletedEvent, event.Type)
		assert.Equal(t, nc2.ID, event.Payload.ID)
	})
}
