package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
	"github.com/leg100/surl"
	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
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
	defer webSrv.Close()

	// setup SSE client and subscribe to stream
	events := make(chan *sse.Event, 1)
	errch := make(chan otf.Event, 1)
	httpClient, err := newTestClient(webSrv.URL)
	require.NoError(t, err)
	client, err := httpClient.newSSEClient("watch", errch)
	require.NoError(t, err)
	require.NoError(t, client.SubscribeChanRaw(events))
	defer client.Unsubscribe(events)

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
	got := <-events
	assert.Equal(t, "run_created", string(got.Event))

	// check event payload is what we expect
	gotRun := dto.Run{}
	err = jsonapi.UnmarshalPayload(bytes.NewReader(got.Data), &gotRun)
	require.NoError(t, err)
	assert.Equal(t, wantRun.ID(), gotRun.ID)
}

type fakeEventsApp struct {
	ch chan otf.Event
	otf.Application
}

func (f *fakeEventsApp) Watch(context.Context, otf.WatchOptions) (<-chan otf.Event, error) {
	return f.ch, nil
}

func (f *fakeEventsApp) ListWorkspacePermissions(context.Context, otf.WorkspaceSpec) ([]*otf.WorkspacePermission, error) {
	return nil, nil
}
