package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

var _ otf.ChunkService = (*logService)(nil)

type logService struct {
	proxy otf.ChunkStore
	db    otf.DB
	logr.Logger
}

func newLogService(db otf.DB, logger logr.Logger, cache otf.Cache) (*logService, error) {
	proxy, err := inmem.NewChunkProxy(cache, db)
	if err != nil {
		return nil, fmt.Errorf("constructing chunk proxy: %w", err)
	}
	return &logService{
		db:     db,
		proxy:  proxy,
		Logger: logger,
	}, nil
}

// GetChunk reads a chunk of logs for a job.
func (s logService) GetChunk(ctx context.Context, jobID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.proxy.GetChunk(ctx, jobID, opts)
	if err != nil {
		s.Error(err, "reading logs", "id", jobID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	s.V(2).Info("read logs", "id", jobID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a job.
func (s logService) PutChunk(ctx context.Context, jobID string, chunk otf.Chunk) error {
	err := s.proxy.PutChunk(ctx, jobID, chunk)
	if err != nil {
		s.Error(err, "writing logs", "id", jobID, "start", chunk.Start, "end", chunk.End)
		return err
	}
	s.V(2).Info("written logs", "id", jobID, "start", chunk.Start, "end", chunk.End)
	return nil
}
