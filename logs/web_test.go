package logs

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
)

func TestTailLogs(t *testing.T) {
	// setup SSE server
	srv := sse.New()
	srv.AutoStream = true
	srv.AutoReplay = false

	// setup logs channel - send a chunk and then close
	chunks := make(chan otf.Chunk, 1)
	svc := &fakeTailService{chunks: chunks}
	handlers := newWebHandlers(svc, logr.Discard())

	r := httptest.NewRequest("", "/?offset=0&stream=tail-123&phase=plan&run_id=run-123", nil)
	w := httptest.NewRecorder()
	handlers.tailRun(w, r)

	// events := make(chan *sse.Event, 99)
	// require.NoError(t, client.SubscribeChan("tail-123", events))
	// defer client.Unsubscribe(events)

	// Client connected to server, now have the tail handler receive an event
	chunks <- otf.Chunk{Data: []byte("some logs")}
	close(chunks)

	// Wait for client to receive relayed event
	//logs := <-events
	//assert.Equal(t, "new-log-chunk", string(logs.Event))
	assert.Equal(t, "{\"html\":\"some logs\\u003cbr\\u003e\",\"offset\":9}", w.Body.String())
	// Closing channel should result in a finished event.
	//finished := <-events
	//assert.Equal(t, "finished", string(finished.Event))
	//assert.Equal(t, "no more logs", string(finished.Data))
}

type fakeTailService struct {
	chunks chan otf.Chunk
}

func (f *fakeTailService) tail(context.Context, otf.GetChunkOptions) (<-chan otf.Chunk, error) {
	return f.chunks, nil
}
