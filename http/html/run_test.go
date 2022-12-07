package html

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRunsHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
	runs := []*otf.Run{
		otf.NewRun(cv, ws, otf.RunCreateOptions{}),
		otf.NewRun(cv, ws, otf.RunCreateOptions{}),
		otf.NewRun(cv, ws, otf.RunCreateOptions{}),
		otf.NewRun(cv, ws, otf.RunCreateOptions{}),
		otf.NewRun(cv, ws, otf.RunCreateOptions{}),
	}
	app := newFakeWebApp(t, &fakeListRunsApp{runs: runs})

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=1&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listRuns(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=2&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listRuns(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=3&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listRuns(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

func TestTailLogs(t *testing.T) {
	run := otf.NewTestRun(t, otf.TestRunCreateOptions{})

	// setup SSE server
	srv := sse.New()
	srv.AutoStream = true
	srv.AutoReplay = false

	// setup logs channel - send a chunk and then close
	chunks := make(chan otf.Chunk, 1)

	// fake app
	app := &Application{
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
	client := sse.NewClient(webSrv.URL + "/" + run.ID() + "?offset=0&stream=tail-123&phase=plan")
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

type fakeListRunsApp struct {
	runs []*otf.Run
	otf.Application
}

func (f *fakeListRunsApp) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      f.runs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.runs)),
	}, nil
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
