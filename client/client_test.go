package client

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/require"
)

func TestWatchClient(t *testing.T) {
	server := sse.New()

	mux := http.NewServeMux()
	mux.HandleFunc("/watch", server.ServeHTTP)
	webSrv := httptest.NewTLSServer(mux)

	config := otfhttp.Config{
		Address:  webSrv.URL,
		Insecure: true,
	}

	// setup client and subscribe to stream
	client, err := New(config)
	require.NoError(t, err)

	got, err := client.Watch(context.Background(), otf.WatchOptions{})
	require.NoError(t, err)

	require.Equal(t, otf.Event{Type: otf.EventInfo, Payload: "successfully connected"}, <-got)

	publishTestRun(t, server, otf.EventRunCreated, "run-123")

	// TODO: test run
}

func publishTestRun(t *testing.T, server *sse.Server, eventType otf.EventType, runID string) {
	var buf bytes.Buffer
	err := jsonapi.MarshalPayloadWithoutIncluded(&buf, &jsonapi.Run{
		ID: runID,
	})
	require.NoError(t, err)
	server.Publish("messages", &sse.Event{
		Event: []byte(eventType),
		Data:  buf.Bytes(),
	})
}
