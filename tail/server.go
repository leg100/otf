package tail

import (
	"context"
	"sync"

	"github.com/leg100/otf"
)

// Server is responsible for tailing logs on behalf of clients
type Server struct {
	// need the service to retrieve any logs persisted after the offset and
	// before the next chunk of logs are uploaded.
	svc otf.ChunkService

	// exclusion lock ensures reads and writes to map are serialized, and also
	// ensures new logs cannot be uploaded whilst a client is being added.
	mu sync.Mutex

	// database of phases and list of clients for each phase
	db map[otf.PhaseSpec][]*client
}

func NewServer(svc otf.ChunkService) *Server {
	return &Server{
		svc: svc,
		db:  make(map[otf.PhaseSpec][]*client),
	}
}

// Tail provides a client that follows logs for the given phase and from the
// given offset onwards.
func (s *Server) Tail(ctx context.Context, spec otf.PhaseSpec, offset int) (<-chan []byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// first retrieve previous logs from db starting at the offset
	prev, err := s.svc.GetChunk(ctx, spec.RunID, spec.Phase, otf.GetChunkOptions{
		Offset: offset,
	})
	if err != nil && err != otf.ErrResourceNotFound {
		return nil, err
	}

	c := &client{
		phase: spec,
		ch:    make(chan []byte, 99999),
	}
	if len(prev.Data) > 0 {
		// send logs from db to client
		c.ch <- prev.Data
	}
	if prev.End {
		// inform client there are no more logs by closing channel
		close(c.ch)
		// we don't need to store client in db
		return c.ch, nil
	}

	// store client in db so that we can send it more logs later
	clients, ok := s.db[spec]
	if !ok {
		// this is the first client tailing this phase
		s.db[spec] = []*client{c}
	} else {
		// there are other clients tailing this phase too
		s.db[spec] = append(clients, c)
	}

	// remove client if it disconnects etc
	go func() {
		<-ctx.Done()
		s.removeClient(c)
	}()

	return c.ch, nil
}

// PutChunk forwards the chunk to both the backend service and to tailing clients.
func (s *Server) PutChunk(ctx context.Context, spec otf.PhaseSpec, chunk otf.Chunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// forward to backend service
	if err := s.svc.PutChunk(ctx, spec.RunID, spec.Phase, chunk); err != nil {
		return err
	}

	clients, ok := s.db[spec]
	if !ok {
		// no clients tailing this run
		return nil
	}
	for _, c := range clients {
		if len(chunk.Data) > 0 {
			c.ch <- chunk.Data
		}
		if chunk.End {
			// inform client this is the last chunk so they know not to tail
			// logs anymore
			close(c.ch)
		}
	}
	if chunk.End {
		// no more logs so remove clients from db
		delete(s.db, spec)
	}
	return nil
}

func (t *Server) removeClient(c *client) {
	t.mu.Lock()
	defer t.mu.Unlock()

	clients, ok := t.db[c.phase]
	if !ok {
		// client is referring to non-existent phase, do nothing
		return
	}
	// find client
	for idx, existing := range clients {
		if existing == c {
			// remove client
			clients = append(clients[:idx], clients[idx+1:]...)
			t.db[c.phase] = clients
			break
		}
	}
	// if last client was deleted then remove entry for this phase
	if len(clients) == 0 {
		delete(t.db, c.phase)
	}
}
