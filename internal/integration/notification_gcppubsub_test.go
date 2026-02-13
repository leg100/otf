package integration

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_NotificationGCPPubSub demonstrates run events triggering the
// sending of notifications to a GCP pub-sub topic.
func TestIntegration_NotificationGCPPubSub(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("gcp pubsub emulator only runs on amd64")
	}
	testutils.SkipIfEnvUnspecified(t, "PUBSUB_EMULATOR_HOST")

	started := time.Now()

	integrationTest(t)

	client, err := pubsub.NewClient(t.Context(), "abc123")
	require.NoError(t, err)

	// topic id must begin with a letter
	topicID := "a" + uuid.NewString()
	topic, err := client.TopicAdminClient.CreateTopic(t.Context(), &pubsubpb.Topic{
		Name: "projects/abc123/topics/" + topicID,
	})
	require.NoError(t, err)
	// sub id must begin with a letter
	subscription, err := client.SubscriptionAdminClient.CreateSubscription(t.Context(), &pubsubpb.Subscription{
		Name:  "projects/abc123/subscriptions/a" + uuid.NewString(),
		Topic: topic.GetName(),
	})
	require.NoError(t, err)
	received := make(chan *pubsub.Message)
	sub := client.Subscriber(subscription.GetName())
	go func() {
		err := sub.Receive(t.Context(), func(_ context.Context, m *pubsub.Message) {
			received <- m
		})
		require.NoError(t, err)
	}()

	daemon, _, ctx := setup(t)

	ws := daemon.createWorkspace(t, ctx, nil)

	// add some tags to the workspace so we can check below that they are added
	// to the pubsub message.
	err = daemon.Workspaces.AddTags(ctx, ws.ID, []workspace.TagSpec{{Name: "foo"}, {Name: "bar"}})
	require.NoError(t, err)

	_, err = daemon.Notifications.Create(ctx, ws.ID, notifications.CreateConfigOptions{
		DestinationType: notifications.DestinationGCPPubSub,
		Enabled:         new(true),
		Name:            new("testing"),
		URL:             new("gcppubsub://abc123/" + topicID),
		Triggers: []notifications.Trigger{
			notifications.TriggerCreated,
			notifications.TriggerPlanning,
			notifications.TriggerNeedsAttention,
		},
	})
	require.NoError(t, err)

	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	run := daemon.createRun(t, ctx, ws, cv, nil)

	// gcp-pubsub messages are not necessarily received in the same order as
	// they are sent, so wait til all expected messages are received and then
	// check them.
	var got []*pubsub.Message
	got = append(got, <-received)
	got = append(got, <-received)
	got = append(got, <-received)
	got = append(got, <-received)

	// keep a record of whether a match was found for each expected status
	matches := map[runstatus.Status]bool{
		runstatus.Pending:  false,
		runstatus.Planning: false,
		runstatus.Planned:  false,
	}
	for _, g := range got {
		var payload notifications.GenericPayload
		err = json.Unmarshal(g.Data, &payload)
		require.NoError(t, err)

		notification := payload.Notifications[0]
		if _, ok := matches[notification.RunStatus]; ok {
			assert.Equal(t, run.ID, payload.RunID)
			matches[notification.RunStatus] = true
		}

		// check attributes include workspace metadata
		want := map[string]string{
			"otf.ninja/v1/workspace.name": ws.Name,
			"otf.ninja/v1/workspace.id":   ws.ID.String(),
			"otf.ninja/v1/tags/foo":       "true",
			"otf.ninja/v1/tags/bar":       "true",
		}
		assert.Equal(t, want, g.Attributes)

		// check time is valid
		assert.True(t, notification.RunUpdatedAt.After(started),
			"time is invalid: %s", notification.RunUpdatedAt.String())

		// check notification includes valid info
		assert.Equal(t, run.ID, payload.RunID)
		assert.Equal(t, run.Organization, payload.OrganizationName)
		assert.Equal(t, ws.Name, payload.WorkspaceName)
	}
	// check statuses were all received
	for status, want := range matches {
		assert.True(t, want, "status %s not received", status)
	}
}
