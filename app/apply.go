package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.ApplyService = (*ApplyService)(nil)

type ApplyService struct {
	db otf.RunStore
	otf.ChunkStore

	logr.Logger
}

func NewApplyService(db otf.RunStore, logs otf.ChunkStore, logger logr.Logger) *ApplyService {
	return &ApplyService{
		db:         db,
		ChunkStore: logs,
		Logger:     logger,
	}
}

func (s ApplyService) Get(id string) (*otf.Apply, error) {
	run, err := s.db.Get(otf.RunGetOptions{ApplyID: &id})
	if err != nil {
		return nil, err
	}
	return run.Apply, nil
}

// GetApplyLogs reads a chunk of logs for a terraform apply.
func (s ApplyService) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) ([]byte, error) {
	logs, err := s.GetChunk(ctx, id, opts)
	if err != nil {
		s.Error(err, "reading apply logs", "id", id, "offset", opts.Offset, "limit", opts.Limit)
		return nil, err
	}

	return logs, nil
}

// UploadLogs writes a chunk of logs for a terraform apply.
func (s ApplyService) PutChunk(ctx context.Context, id string, chunk []byte, opts otf.PutChunkOptions) error {
	err := s.PutChunk(ctx, id, chunk, opts)
	if err != nil {
		s.Error(err, "writing apply logs", "id", id, "start", opts.Start, "end", opts.End)
		return err
	}

	if !opts.End {
		return nil
	}

	// Last chunk uploaded. A summary of applied changes can now be parsed from
	// the full logs and set on the apply obj.
	logs, err := s.GetChunk(ctx, id, otf.GetChunkOptions{})
	if err != nil {
		s.Error(err, "reading apply logs", "id", id)
		return err
	}

	_, err = s.db.Update(otf.RunUpdateOptions{ApplyID: otf.String(id)}, func(run *otf.Run) (err error) {
		run.Apply.Resources, err = otf.ParseApplyOutput(string(logs))

		return err
	})
	if err != nil {
		s.Error(err, "summarising applied changes", "id", id)
		return err
	}

	return nil
}
