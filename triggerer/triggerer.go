// Package triggerer handles triggering things in response to incoming VCS
// events.
package triggerer

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/module"
)

// Triggerer triggers jobs in response to incoming VCS events.
type Triggerer struct {
	otf.Application
	*module.Publisher
	logr.Logger
}

func NewTriggerer(app otf.Application, logger logr.Logger) *Triggerer {
	return &Triggerer{
		Application: app,
		Publisher:   module.NewPublisher(app),
		Logger:      logger.WithValues("component", "triggerer"),
	}
}

// Start handling VCS events and triggering jobs
func (h *Triggerer) Start(ctx context.Context) error {
	h.V(2).Info("started")

	sub, err := h.Subscribe(ctx, "triggerer")
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-sub:
			// skip non-vcs events
			if event.Type != otf.EventVCS {
				continue
			}

			if err := h.handle(ctx, event.Payload); err != nil {
				h.Error(err, "handling vcs event")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handle triggers a run upon receiving an event
func (h *Triggerer) handle(ctx context.Context, event cloud.VCSEvent) error {
	if err := h.triggerRun(ctx, event); err != nil {
		return err
	}

	if err := h.PublishFromEvent(ctx, event); err != nil {
		return err
	}

	return nil
}
