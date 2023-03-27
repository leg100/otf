package run

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
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
		var buf bytes.Buffer
		err := jsonapi.MarshalPayloadWithoutIncluded(&buf, &jsonapi.Run{
			ID:                   "run-123",
			Workspace:            &jsonapi.Workspace{ID: "ws-123"},
			ConfigurationVersion: &jsonapi.ConfigurationVersion{ID: "cv-123"},
		})
		require.NoError(t, err)
		otf.WriteSSEEvent(w, buf.Bytes(), otf.EventRunStatusUpdate, true)
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

	assert.Equal(t, otf.Event{Type: otf.EventInfo, Payload: "successfully connected"}, <-got)
	want := &Run{
		ID:                     "run-123",
		WorkspaceID:            "ws-123",
		ConfigurationVersionID: "cv-123",
	}
	assert.Equal(t, otf.Event{Type: otf.EventRunStatusUpdate, Payload: want}, <-got)
}
