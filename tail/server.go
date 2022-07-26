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
	db map[otf.PhaseSpec][]*Client
}

func NewServer(svc otf.ChunkService) *Server {
	return &Server{
		svc: svc,
		db:  make(map[otf.PhaseSpec][]*Client),
	}
}

// Tail provides a client that follows logs for the given phase and from the
// given offset onwards.
func (s *Server) Tail(ctx context.Context, spec otf.PhaseSpec, offset int) (*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// first retrieve previous logs from db starting at the offset
	prev, err := s.svc.GetChunk(ctx, spec.RunID, spec.Phase, otf.GetChunkOptions{
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	client := &Client{
		server: s,
		phase:  spec,
		buffer: make(chan []byte, 99999),
	}
	// send logs from db to client
	if len(prev.Data) > 0 {
		client.buffer <- prev.Data
	}
	// and inform client there are no more logs by closing channel
	if prev.End {
		close(client.buffer)
		// we don't need to store client in db
		return client, nil
	}

	// store client in db so that we can send it more logs later
	clients, ok := s.db[spec]
	if !ok {
		// this is the first client tailing this phase
		s.db[spec] = []*Client{client}
	} else {
		// there are other clients tailing this phase too
		s.db[spec] = append(clients, client)
	}

	return client, nil
}

// PutChunk should be called whenever a chunk of logs is written to the system,
// so that clients receive a copy.
func (t *Server) PutChunk(spec otf.PhaseSpec, chunk otf.Chunk) {
	t.mu.Lock()
	defer t.mu.Unlock()

	clients, ok := t.db[spec]
	if !ok {
		// no clients tailing this run
		return
	}
	for _, c := range clients {
		c.buffer <- chunk.Data
		if chunk.End {
			// inform clients this is the last chunk so they know not to tail
			// logs anymore
			close(c.buffer)
		}
	}
}

func (t *Server) removeClient(client *Client) {
	t.mu.Lock()
	defer t.mu.Unlock()

	clients, ok := t.db[client.phase]
	if !ok {
		// client is referring to non-existent phase, do nothing
		return
	}
	// find client
	for idx, existing := range clients {
		if existing == client {
			// remove client
			clients = append(clients[:idx], clients[idx+1:]...)
			t.db[client.phase] = clients
			break
		}
	}
	// if last client was deleted then remove entry for this phase
	if len(clients) == 0 {
		delete(t.db, client.phase)
	}
}
