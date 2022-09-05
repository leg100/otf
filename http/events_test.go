package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
	// fake run event
	run := otf.NewTestRun(t, otf.TestRunCreateOptions{})
	want := otf.Event{Type: otf.EventRunCreated, Payload: run}

	// setup SSE server
	eventsServer := sse.New()
	eventsServer.EncodeBase64 = true

	// fake server object
	srv := &Server{
		Application:  newFakeEventServer(want),
		Logger:       logr.Discard(),
		eventsServer: eventsServer,
	}

	// setup web server
	router := mux.NewRouter()
	router.HandleFunc("/watch", srv.watch)
	webSrv := httptest.NewTLSServer(router)
	defer webSrv.Close()

	// setup SSE client and subscribe to stream
	client := sse.NewClient(webSrv.URL + "/watch")
	client.EncodingBase64 = true
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	events := make(chan *sse.Event, 1)
	require.NoError(t, client.SubscribeChanRaw(events))
	defer client.Unsubscribe(events)

	// check event type is what we expect
	got := <-events
	assert.Equal(t, "run_created", string(got.Event))

	// check event payload is what we expect
	dto := dto.Run{}
	err := jsonapi.UnmarshalPayload(bytes.NewReader(got.Data), &dto)
	require.NoError(t, err)
	assert.Equal(t, run.ID(), dto.ID)
}

type fakeEventsApp struct {
	events chan otf.Event
	otf.Application
}

func newFakeEventServer(ev otf.Event) *fakeEventsApp {
	// setup events channel - send an event and then close
	events := make(chan otf.Event, 1)
	events <- ev
	close(events)
	return &fakeEventsApp{events: events}
}

func (f *fakeEventsApp) Watch(context.Context, otf.WatchOptions) (<-chan otf.Event, error) {
	return f.events, nil
}
