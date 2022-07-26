package html

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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

	// fake app
	app := &Application{
		Server: sse.New(),
		Application: &fakeTailApp{
			run:    run,
			client: &fakeTailClient{chunk},
		},
	}

	// setup req and resp
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	mux.SetURLVars(r, map[string]string{"run_id": run.ID()})
	r.URL.RawQuery = url.Values{
		"offset": []string{"0"},
		"stream": []string{"tail-123"},
	}.Encode()

	// run handler in bg and wait for it to respond
	go app.tailPhase("plan")(w, r)

	time.Sleep(time.Second)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "id: 0\ndata: some logs<br>\n\n", w.Body.String())
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
