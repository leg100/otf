package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.LogService = (*LogService)(nil)

type LogService struct {
	PlanLogStore  otf.ChunkStore
	ApplyLogStore otf.ChunkStore

	logr.Logger
}

func NewLogService(planLogsDB, applyLogsDB otf.ChunkStore, logger logr.Logger) *LogService {
	return &LogService{
		PlanLogStore:  planLogsDB,
		ApplyLogStore: applyLogsDB,
		Logger:        logger,
	}
}

func (s LogService) GetPlanLogs(ctx context.Context, planID string, opts otf.GetChunkOptions) ([]byte, error) {
	chunk, err := s.PlanLogStore.GetChunk(planID, opts)
	if err != nil {
		s.Error(err, "reading plan logs", "plan_id", planID, "offset", opts.Offset, "limit", opts.Limit)
		return nil, err
	}

	return chunk, nil
}

func (s LogService) GetApplyLogs(ctx context.Context, applyID string, opts otf.GetChunkOptions) ([]byte, error) {
	chunk, err := s.ApplyLogStore.GetChunk(applyID, opts)
	if err != nil {
		s.Error(err, "reading apply logs", "apply_id", applyID, "offset", opts.Offset, "limit", opts.Limit)
		return nil, err
	}

	return chunk, nil
}

func (s LogService) PutPlanLogs(ctx context.Context, planID string, chunk []byte, opts otf.PutChunkOptions) ([]byte, error) {
	err := s.PlanLogStore.PutChunk(planID, chunk, opts)
	if err != nil {
		s.Error(err, "writing plan logs", "plan_id", planID, "start", opts.Start, "end", opts.End)
		return nil, err
	}

	return chunk, nil
}

func (s LogService) PutApplyLogs(ctx context.Context, applyID string, chunk []byte, opts otf.PutChunkOptions) ([]byte, error) {
	err := s.ApplyLogStore.PutChunk(applyID, chunk, opts)
	if err != nil {
		s.Error(err, "writing apply logs", "apply_id", applyID, "start", opts.Start, "end", opts.End)
		return nil, err
	}

	return chunk, nil
}
