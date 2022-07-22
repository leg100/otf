package app

import (
	"io"
	"sync"

	"github.com/leg100/otf"
)

type tailKey struct {
	runID string
	phase string
}

// Rename to tailServer?
//
// tailer is responsible for tailing logs on behalf of clients
type tailer struct {
	svc otf.ChunkService

	mu sync.Mutex
	// TODO: allocate map in constructor
	//
	// mapping of a run phase to a list of buffers, one for each client
	db map[tailKey][]*tailBuffer
}

// tail returns a buffer of logs for a run and phase from the given offset. The
// buffer will
func (t *tailer) tail(runID string, phase string, offset int) *tailBuffer {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := tailKey{runID, phase}

	// first retrieve logs from db starting at the offset

	// provision buffer for client
	buf := &tailBuffer{
		t:      t,
		phase:  key,
		buffer: make([]byte, 0),
	}

	buffers, ok := t.db[key]
	if !ok {
		// this is the first client tailing this phase
		t.db[key] = []*tailBuffer{buf}
	} else {
		// there are other clients tailing this phase too
		t.db[key] = append(buffers, buf)
	}

	return buf
}

func (t *tailer) removeClient(client *tailBuffer) {
	t.mu.Lock()
	defer t.mu.Unlock()

	buffers, ok := t.db[client.phase]
	if !ok {
		// client is referring to non-existent phase, do nothing
		return
	}
	// find client
	for idx, existing := range buffers {
		if existing == client {
			// remove client
			buffers = append(buffers[:idx], buffers[idx+1:]...)
			t.db[client.phase] = buffers
			break
		}
	}
	// if last client was deleted then remove entry for this phase
	if len(buffers) == 0 {
		delete(t.db, client.phase)
	}
}

func (t *tailer) putChunk(runID, phase string, chunk otf.Chunk) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := tailKey{runID, phase}

	buffers, ok := t.db[key]
	if !ok {
		// no clients tailing this run
		return
	}
	for _, buf := range buffers {
		buf.buffer = append(buf.buffer, chunk.Data...)
		if chunk.End {
			buf.finished = true
		}
	}
}

// Rename to tailClient?
//
// tailBuffer is the buffer of logs for a client. The buffer is written to by
// tailer and read from by the client. The j
type tailBuffer struct {
	finished bool
	t        *tailer
	phase    tailKey
	buffer   []byte
}

func (t *tailBuffer) Read(p []byte) (int, error) {
	read := copy(p, t.buffer)
	if t.finished {
		return read, io.EOF
	} else {
		return read, nil
	}
}

func (t *tailBuffer) Close() {
	t.t.removeClient(t)
}
