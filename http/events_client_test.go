package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/surl"
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
		Signer:       surl.New([]byte("secretsauce")),
	}

	// setup web server
	router := mux.NewRouter()
	// adds a subject with unlimited privs so we can by-pass authz
	router.Handle("/watch", allowAllMiddleware(http.HandlerFunc(srv.watch)))
	webSrv := httptest.NewTLSServer(router)
	t.Cleanup(func() {
		// closing chan terminates conn allowing server to close without timing
		// out
		close(serverCh)
		defer webSrv.Close()
	})

	// setup client and subscribe to stream
	client, err := newTestClient(webSrv.URL)
	require.NoError(t, err)
	clientCh, err := client.Watch(context.Background(), otf.WatchOptions{})
	require.NoError(t, err)

	require.Equal(t, otf.Event{Type: otf.EventInfo, Payload: "successfully connected"}, <-clientCh)

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
}
