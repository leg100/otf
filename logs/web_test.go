package logs

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTailLogs(t *testing.T) {
	run := otf.NewTestRun(t, otf.TestRunCreateOptions{})

	// setup SSE server
	srv := sse.New()
	srv.AutoStream = true
	srv.AutoReplay = false

	// setup logs channel - send a chunk and then close
	chunks := make(chan otf.Chunk, 1)

	// fake app
	app := &app{
		Server: srv,
		Application: &fakeTailApp{
			run:    run,
			chunks: chunks,
		},
		Logger: logr.Discard(),
	}

	// setup web server
	router := mux.NewRouter()
	router.HandleFunc("/{run_id}", app.tailRun)
	webSrv := httptest.NewTLSServer(router)
	defer webSrv.Close()

	// setup SSE client and subscribe to stream
	client := sse.NewClient(webSrv.URL + "/" + run.ID + "?offset=0&stream=tail-123&phase=plan")
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	events := make(chan *sse.Event, 99)
	require.NoError(t, client.SubscribeChan("tail-123", events))
	defer client.Unsubscribe(events)

	// Client connected to server, now have the tail handler receive an event
	chunks <- otf.Chunk{Data: []byte("some logs")}
	close(chunks)

	// Wait for client to receive relayed event
	logs := <-events
	assert.Equal(t, "new-log-chunk", string(logs.Event))
	assert.Equal(t, "{\"html\":\"some logs\\u003cbr\\u003e\",\"offset\":9}", string(logs.Data))
	// Closing channel should result in a finished event.
	finished := <-events
	assert.Equal(t, "finished", string(finished.Event))
	assert.Equal(t, "no more logs", string(finished.Data))
}

type fakeTailApp struct {
	run    *otf.Run
	chunks chan otf.Chunk

	otf.Application
}

func (f *fakeTailApp) GetRun(context.Context, string) (*otf.Run, error) {
	return f.run, nil
}

func (f *fakeTailApp) Tail(context.Context, otf.GetChunkOptions) (<-chan otf.Chunk, error) {
	return f.chunks, nil
}
