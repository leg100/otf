package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
)

var _ otf.ApplyService = (*ApplyService)(nil)

type ApplyService struct {
	db *sql.DB
	cs otf.ChunkService
	logr.Logger
}

func NewApplyService(db *sql.DB, logger logr.Logger, cs otf.ChunkService) *ApplyService {
	return &ApplyService{
		cs:     cs,
		db:     db,
		Logger: logger,
	}
}

func (s ApplyService) Get(ctx context.Context, id string) (*otf.Apply, error) {
	run, err := s.db.GetRun(ctx, otf.RunGetOptions{ApplyID: &id})
	if err != nil {
		return nil, err
	}
	return run.Apply, nil
}
