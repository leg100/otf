package http

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchClient(t *testing.T) {
	// fake server
	serverCh := make(chan otf.Event, 1)
	srv := &Server{
		Application:  &fakeEventsApp{ch: serverCh},
		Logger:       logr.Discard(),
		eventsServer: newSSEServer(),
	}

	// setup web server
	router := mux.NewRouter()
	router.HandleFunc("/watch", srv.watch)
	webSrv := httptest.NewTLSServer(router)
	defer webSrv.Close()

	// setup client and subscribe to stream
	client, err := newTestClient(webSrv.URL)
	require.NoError(t, err)
	clientCh, err := client.Watch(context.Background(), otf.WatchOptions{})
	require.NoError(t, err)

	// Give client time to connect and subscribe before publishing message
	time.Sleep(100 * time.Millisecond)

	// publish message server-side
	wantRun := otf.NewTestRun(t, otf.TestRunCreateOptions{})
	ev := otf.Event{
		Type:    otf.EventRunCreated,
		Payload: wantRun,
	}
	serverCh <- ev

	// check received event type is what we expect
	got := <-clientCh
	assert.Equal(t, otf.EventRunCreated, otf.EventType(got.Type))

	// check event payload is what we expect
	gotRun, ok := got.Payload.(*otf.Run)
	assert.True(t, ok)
	assert.Equal(t, wantRun.ID(), gotRun.ID())
	assert.Equal(t, wantRun.Status(), gotRun.Status())
	assert.Equal(t, wantRun.ConfigurationVersionID(), gotRun.ConfigurationVersionID())
	assert.Equal(t, wantRun.WorkspaceID(), gotRun.WorkspaceID())

	// closing server chan terminates connection from the server-side,
	// allowing the temp web server's deferred close above to complete,
	// otherwise the test times out.
	close(serverCh)
}
