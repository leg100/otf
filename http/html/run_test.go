package html

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
)

func Test_TailLogs(t *testing.T) {
	run := otf.NewTestRun(t, otf.TestRunCreateOptions{})

	// setup channel - send one chunk and then close
	chunk := make(chan []byte, 1000)
	chunk <- []byte("some logs")
	close(chunk)

	// setup SSE server
	srv := sse.New()
	srv.AutoStream = true
	srv.AutoReplay = false

	// fake app
	app := &Application{
		Server: srv,
		Application: &fakeTailApp{
			run:    run,
			client: &fakeTailClient{chunk},
		},
		Logger: logr.Discard(),
	}

	// setup req and resp
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	mux.SetURLVars(r, map[string]string{"run_id": run.ID()})
	r.URL.RawQuery = url.Values{
		"offset": []string{"0"},
		"stream": []string{"tail-123"},
		"phase":  []string{"plan"},
	}.Encode()

	// run handler in bg and wait for it to respond
	go app.tailRun(w, r)

	time.Sleep(time.Second)

	assert.Equal(t, 200, w.Code)
	want := `id: 
data: {"offset":9,"html":"some logs\u003cbr\u003e"}
event: new-log-chunk

id: 
data: no more logs
event: finished

`
	assert.Equal(t, want, w.Body.String())
}

type fakeTailApp struct {
	run    *otf.Run
	client *fakeTailClient

	otf.Application
}

func (f *fakeTailApp) GetRun(context.Context, string) (*otf.Run, error) {
	return f.run, nil
}

func (f *fakeTailApp) Tail(context.Context, string, otf.PhaseType, int) (otf.TailClient, error) {
	return f.client, nil
}

type fakeTailClient struct {
	chunk chan []byte
}

func (f *fakeTailClient) Read() <-chan []byte {
	return f.chunk
}

func (f *fakeTailClient) Close() {}
