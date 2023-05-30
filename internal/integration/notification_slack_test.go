package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/notifications"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_NotificationSlack demonstrates run events triggering the
// sending of notifications to a slack channel.
func TestIntegration_NotificationSlack(t *testing.T) {
	t.Parallel()

	url, got := newSlackServer(t)
	daemon := setup(t, nil)

	ws := daemon.createWorkspace(t, ctx, nil)
	_, err := daemon.CreateNotificationConfiguration(ctx, ws.ID, notifications.CreateConfigOptions{
		DestinationType: notifications.DestinationSlack,
		Enabled:         internal.Bool(true),
		Name:            internal.String("testing"),
		URL:             internal.String(url),
		Triggers: []notifications.Trigger{
			notifications.TriggerCreated,
			notifications.TriggerPlanning,
			notifications.TriggerNeedsAttention,
		},
	})
	require.NoError(t, err)

	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	_ = daemon.createRun(t, ctx, ws, cv)

	assert.Regexp(t, `run pending`, <-got)
	assert.Regexp(t, `run planning`, <-got)
	assert.Regexp(t, `run planning`, <-got)
	assert.Regexp(t, `run planned`, <-got)
}

func newSlackServer(t *testing.T) (string, <-chan string) {
	got := make(chan string, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/mychannel", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got <- string(body)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL + "/mychannel", got
}
