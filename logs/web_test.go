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

	r := httptest.NewRequest("", "/?offset=0&stream=tail-123&phase=plan&run_id=run-123", nil)
	w := httptest.NewRecorder()

	// send one event and then close.
	chunks <- otf.Chunk{Data: []byte("some logs")}
	close(chunks)

	done := make(chan struct{})
	go func() {
		handlers.tailRun(w, r)

		// test chunk received
		want := `data: {"html":"some logs\u003cbr\u003e","offset":9}
event: new-log-chunk

event: finished
data: no more logs

`
		assert.Equal(t, want, w.Body.String())

		done <- struct{}{}
	}()
	<-done
}

type fakeTailService struct {
	chunks chan otf.Chunk
}

func (f *fakeTailService) tail(context.Context, otf.GetChunkOptions) (<-chan otf.Chunk, error) {
	return f.chunks, nil
}
