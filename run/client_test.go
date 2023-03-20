package run

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchClient(t *testing.T) {
	server := sse.New()
	server.AutoStream = true
	server.EncodeBase64 = true

	mux := http.NewServeMux()
	mux.HandleFunc("/watch", func(w http.ResponseWriter, r *http.Request) {
		// r3's sse server expects a stream parameter to be set
		q := r.URL.Query()
		q.Add("stream", "messages")
		r.URL.RawQuery = q.Encode()
		server.ServeHTTP(w, r)
	})
	webserver := httptest.NewTLSServer(mux)
	defer server.Close()

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
	publishTestRun(t, server, otf.EventRunStatusUpdate, "run-123")
	assert.Equal(t, otf.Event{Type: otf.EventRunStatusUpdate, Payload: &Run{ID: "run-123"}}, <-got)
}

func publishTestRun(t *testing.T, server *sse.Server, eventType otf.EventType, runID string) {
	data, err := json.Marshal(&Run{
		ID: runID,
	})
	require.NoError(t, err)
	server.Publish("messages", &sse.Event{
		Event: []byte(eventType),
		Data:  data,
	})
}
