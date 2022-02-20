package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

type runCreator struct {
	db otf.RunStore
	es otf.EventService

	logr.Logger
}

func (s runCreator) createRun(ctx context.Context, run *otf.Run) (*otf.Run, error) {
	_, err := s.db.Create(run)
	if err != nil {
		s.Error(err, "creating run", "id", run.ID)
		return nil, err
	}

	s.V(1).Info("created run", "id", run.ID)

	s.es.Publish(otf.Event{Type: otf.EventRunCreated, Payload: run})

	return run, nil
}
