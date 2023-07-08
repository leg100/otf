package run

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchClient(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/watch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "\r\n")
		b, err := jsonapi.Marshal(&types.Run{
			ID:                   "run-123",
			Workspace:            &types.Workspace{ID: "ws-123"},
			ConfigurationVersion: &types.ConfigurationVersion{ID: "cv-123"},
		})
		require.NoError(t, err)
		pubsub.WriteSSEEvent(w, b, pubsub.UpdatedEvent, true)
	})
	webserver := httptest.NewTLSServer(mux)

	// setup client and subscribe to stream
	client := &Client{
		Config: otfhttp.Config{
			Address:  webserver.URL,
			Insecure: true,
		},
	}

	got, err := client.Watch(context.Background(), WatchOptions{})
	require.NoError(t, err)

	assert.Equal(t, pubsub.Event{Type: pubsub.EventInfo, Payload: "successfully connected"}, <-got)
	want := &Run{
		ID:                     "run-123",
		WorkspaceID:            "ws-123",
		ConfigurationVersionID: "cv-123",
	}
	assert.Equal(t, pubsub.Event{Type: pubsub.UpdatedEvent, Payload: want}, <-got)
}
