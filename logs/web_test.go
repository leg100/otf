package logs

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestTailLogs(t *testing.T) {
	chunks := make(chan otf.Chunk, 1)
	handlers := &webHandlers{
		Logger: logr.Discard(),
		svc:    &fakeTailService{chunks: chunks},
	}

	r := httptest.NewRequest("", "/?offset=0&phase=plan&run_id=run-123", nil)
	w := httptest.NewRecorder()

	// send one event and then close.
	chunks <- otf.Chunk{Data: []byte("some logs")}
	close(chunks)

	done := make(chan struct{})
	go func() {
		handlers.tailRun(w, r)

		// should receive base64 encoded event
		want := "data: eyJodG1sIjoic29tZSBsb2dzXHUwMDNjYnJcdTAwM2UiLCJvZmZzZXQiOjl9\nevent: log_update\n\ndata: bm8gbW9yZSBsb2dz\nevent: log_finished\n\n"
		assert.Equal(t, want, w.Body.String())

		done <- struct{}{}
	}()
	<-done
}

type fakeTailService struct {
	chunks chan otf.Chunk
}

func (f *fakeTailService) Tail(context.Context, otf.GetChunkOptions) (<-chan otf.Chunk, error) {
	return f.chunks, nil
}
