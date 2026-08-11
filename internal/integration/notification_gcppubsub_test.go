package integration

import (
	"context"
	"encoding/json"
	"errors"
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
	// Ack messages, otherwise they are redelivered, and buffer them so the
	// callback doesn't block once the test stops reading.
	received := make(chan *pubsub.Message, 100)
	sub := client.Subscriber(subscription.GetName())
	go func() {
		err := sub.Receive(t.Context(), func(_ context.Context, m *pubsub.Message) {
			m.Ack()
			select {
			case received <- m:
			default:
			}
		})
		// Cancellation on test teardown is expected. The test cannot be failed
		// from here in any case: wrong goroutine.
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Logf("subscriber stopped: %s", err)
		}
	}()

	daemon, _, ctx := setup(t)

	ws := daemon.createWorkspace(t, ctx, nil)

	// add some tags to the workspace so we can check below that they are added
	// to the pubsub message.
	err = daemon.Workspaces.AddTags(ctx, ws.ID, []workspace.TagSpec{{Name: "foo"}, {Name: "bar"}})
	require.NoError(t, err)

	_, err = daemon.Notifications.CreateNotificationConfig(ctx, ws.ID, notifications.CreateConfigOptions{
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

	// keep a record of whether a match was found for each expected status
	matches := map[runstatus.Status]bool{
		runstatus.Pending:  false,
		runstatus.Planning: false,
		runstatus.Planned:  false,
	}
	missing := func() []runstatus.Status {
		var statuses []runstatus.Status
		for status, seen := range matches {
			if !seen {
				statuses = append(statuses, status)
			}
		}
		return statuses
	}

	// Wait for each expected status. Notifications are published per run event
	// rather than per status transition, so neither the number of messages nor
	// their order is fixed. Bounded, so a missing status fails this test rather
	// than the whole package's timeout.
	deadline := time.After(time.Minute)
receive:
	for len(missing()) > 0 {
		var g *pubsub.Message
		select {
		case <-deadline:
			// Include the run status: a failed run sends no notification, so
			// otherwise an errored plan is indistinguishable from a slow one. Not
			// via getRun, which is fatal on error.
			var status runstatus.Status
			if r, rerr := daemon.Runs.GetRun(ctx, run.ID); rerr == nil {
				status = r.Status
			}
			t.Errorf("timed out waiting for notifications; statuses not received: %v; run status: %s",
				missing(), status)
			break receive
		case g = <-received:
		}

		var payload notifications.GenericPayload
		err = json.Unmarshal(g.Data, &payload)
		require.NoError(t, err)

		notification := payload.Notifications[0]
		if _, ok := matches[notification.RunStatus]; ok {
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
}
