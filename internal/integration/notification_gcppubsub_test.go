package integration

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/notifications"
	otfrun "github.com/leg100/otf/internal/run"
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

	integrationTest(t)
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, "abc123")
	require.NoError(t, err)
	defer client.Close()

	// topic id must begin with a letter
	topic, err := client.CreateTopic(ctx, "a"+uuid.NewString())
	require.NoError(t, err)
	// sub id must begin with a letter
	sub, err := client.CreateSubscription(ctx, "a"+uuid.NewString(), pubsub.SubscriptionConfig{
		Topic: topic,
	})
	require.NoError(t, err)
	received := make(chan *pubsub.Message)
	go func() {
		sub.Receive(ctx, func(_ context.Context, m *pubsub.Message) {
			received <- m
		})
	}()

	daemon, _, ctx := setup(t, nil)

	ws := daemon.createWorkspace(t, ctx, nil)

	// add some tags to the workspace so we can check below that they are added
	// to the pubsub message.
	err = daemon.Workspaces.AddTags(ctx, ws.ID, []workspace.TagSpec{{Name: "foo"}, {Name: "bar"}})
	require.NoError(t, err)

	_, err = daemon.Notifications.Create(ctx, ws.ID, notifications.CreateConfigOptions{
		DestinationType: notifications.DestinationGCPPubSub,
		Enabled:         internal.Bool(true),
		Name:            internal.String("testing"),
		URL:             internal.String("gcppubsub://abc123/" + topic.ID()),
		Triggers: []notifications.Trigger{
			notifications.TriggerCreated,
			notifications.TriggerPlanning,
			notifications.TriggerNeedsAttention,
		},
	})
	require.NoError(t, err)

	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	run := daemon.createRun(t, ctx, ws, cv)

	// gcp-pubsub messages are not necessarily received in the same order as
	// they are sent, so wait til all expected messages are received and then
	// check them.
	var got []*pubsub.Message
	got = append(got, <-received)
	got = append(got, <-received)
	got = append(got, <-received)
	got = append(got, <-received)

	// keep a record of whether a match was found for each expected status
	matches := map[otfrun.Status]bool{
		otfrun.RunPending:  false,
		otfrun.RunPlanning: false,
		otfrun.RunPlanned:  false,
	}
	for _, g := range got {
		var payload notifications.GenericPayload
		err = json.Unmarshal(g.Data, &payload)
		require.NoError(t, err)
		status := payload.Notifications[0].RunStatus
		if _, ok := matches[status]; ok {
			assert.Equal(t, run.ID, payload.RunID)
			matches[status] = true
		}
		// check attributes include workspace metadata
		want := map[string]string{
			"otf.ninja/v1/workspace.name": ws.Name,
			"otf.ninja/v1/workspace.id":   ws.ID,
			"otf.ninja/v1/tags/foo":       "true",
			"otf.ninja/v1/tags/bar":       "true",
		}
		assert.Equal(t, want, g.Attributes)
	}
	// check everything matched
	for _, want := range matches {
		assert.True(t, want)
	}
}
