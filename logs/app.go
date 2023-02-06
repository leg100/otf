package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type Application struct {
	otf.Authorizer // authorize access
	logr.Logger

	proxy ChunkProxy
	db
}

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (a *Application) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	logs, err := a.proxy.GetChunk(ctx, opts)
	if err == otf.ErrResourceNotFound {
		// ignore resource not found because no log chunks may not have been
		// written yet
		return Chunk{}, nil
	} else if err != nil {
		a.Error(err, "reading logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
		return Chunk{}, err
	}
	a.V(2).Info("read logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a phase.
func (a *Application) PutChunk(ctx context.Context, chunk Chunk) error {
	_, err := a.CanAccessRun(ctx, rbac.PutChunkAction, chunk.RunID)
	if err != nil {
		return err
	}

	persisted, err := a.db.PutChunk(ctx, chunk)
	if err != nil {
		a.Error(err, "writing logs", "id", chunk.RunID, "phase", chunk.Phase, "offset", chunk.Offset)
		return err
	}
	a.V(2).Info("written logs", "id", chunk.RunID, "phase", chunk.Phase, "offset", chunk.Offset)

	a.Publish(otf.Event{
		Type:    otf.EventLogChunk,
		Payload: persisted,
	})

	return nil
}
