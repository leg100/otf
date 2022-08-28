package tail

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
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

	logger logr.Logger
}

func NewServer(svc otf.ChunkService, logger logr.Logger) *Server {
	return &Server{
		svc:    svc,
		db:     make(map[otf.PhaseSpec][]*Client),
		logger: logger,
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
	if err != nil && err != otf.ErrResourceNotFound {
		return nil, err
	}

	client := &Client{
		server: s,
		phase:  spec,
		buffer: make(chan []byte, 99999),
	}
	if len(prev.Data) > 0 {
		// send logs from db to client
		client.buffer <- prev.Data
		s.logger.Info("sending interim chunk to client", "data", string(prev.Data))
	}
	if prev.End {
		// inform client there are no more logs by closing channel
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
			s.logger.Info("sending chunk to client", "data", string(chunk.Data), "client", c)
			c.buffer <- chunk.Data
		}
		if chunk.End {
			// inform client this is the last chunk so they know not to tail
			// logs anymore
			close(c.buffer)
		}
	}
	return nil
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
