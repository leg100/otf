package logs

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

func TestTailLogs(t *testing.T) {
	chunks := make(chan Chunk, 1)
	handlers := &webHandlers{
		Logger: logr.Discard(),
		svc:    &fakeTailService{chunks: chunks},
	}

	r := httptest.NewRequest("", "/?offset=0&phase=plan&run_id=run-123", nil)
	w := httptest.NewRecorder()

	// send one event and then close.
	chunks <- Chunk{Data: []byte("some logs")}
	close(chunks)

	done := make(chan struct{})
	go func() {
		handlers.tailRun(w, r)

		// should receive base64 encoded event
		want := "data: {\"html\":\"some logs\\u003cbr\\u003e\",\"offset\":9}\nevent: log_update\n\ndata: no more logs\nevent: log_finished\n\n"
		assert.Equal(t, want, w.Body.String())

		done <- struct{}{}
	}()
	<-done
}

type fakeTailService struct {
	chunks chan Chunk
}

func (f *fakeTailService) Tail(context.Context, TailOptions) (<-chan Chunk, error) {
	return f.chunks, nil
}
