package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.RunCreator = (*RunCreator)(nil)

type RunCreator struct {
	db otf.RunStore
	es otf.EventService

	logr.Logger
}

func NewRunCreator(db otf.RunStore, logger logr.Logger, es otf.EventService) RunCreator {
	return RunCreator{
		db:     db,
		es:     es,
		Logger: logger,
	}
}

func (s RunCreator) CreateRun(ctx context.Context, run *otf.Run) (*otf.Run, error) {
	_, err := s.db.Create(run)
	if err != nil {
		s.Error(err, "creating run", "id", run.ID)
		return nil, err
	}

	s.V(1).Info("created run", "id", run.ID)

	s.es.Publish(otf.Event{Type: otf.EventRunCreated, Payload: run})

	return run, nil
}
