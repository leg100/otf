package http

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchClient(t *testing.T) {
	// fake run event
	want := otf.Event{
		Type:    otf.EventRunCreated,
		Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{}),
	}

	// fake server object
	srv := &Server{
		Application:  newFakeEventServer(want),
		Logger:       logr.Discard(),
		eventsServer: newSSEServer(),
	}

	// setup web server
	router := mux.NewRouter()
	router.HandleFunc("/watch", srv.watch)
	webSrv := httptest.NewTLSServer(router)
	defer webSrv.Close()
	u, err := url.Parse(webSrv.URL)
	require.NoError(t, err)

	// setup client and subscribe to stream
	client := client{
		baseURL:  u,
		insecure: true,
	}
	events, err := client.Watch(context.Background(), otf.WatchOptions{})
	require.NoError(t, err)

	// check event type is what we expect
	got := <-events
	assert.Equal(t, otf.EventRunCreated, otf.EventType(got.Type))

	// check event payload is what we expect
	run, ok := got.Payload.(*otf.Run)
	assert.True(t, ok)
	assert.Equal(t, want.Payload.(*otf.Run).ID(), run.ID())
	assert.Equal(t, want.Payload.(*otf.Run).Status(), run.Status())
	assert.Equal(t, want.Payload.(*otf.Run).ConfigurationVersionID(), run.ConfigurationVersionID())
	assert.Equal(t, want.Payload.(*otf.Run).WorkspaceID(), run.WorkspaceID())
}
