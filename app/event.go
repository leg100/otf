package app

import (
	"context"

	"github.com/leg100/otf"
)

// OrganizationResource is a resource that belongs to an organization
type OrganizationResource interface {
	OrganizationName() string
}

// Watch provides authenticated access to a stream of events.
//
// TODO: apply watch options
func (a *Application) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	ch := make(chan otf.Event)
	go func() {
		sub := a.Subscribe(ctx)
		for {
			select {
			case ev, ok := <-sub:
				res, ok := ev.Payload.(OrganizationResource)
				if !ok {
					// skip events that contain payloads that cannot be related
					// back to an organization
					continue
				}
				if !otf.CanAccess(ctx, otf.String(res.OrganizationName())) {
					// skip events caller is not entitled to
					continue
				}
				ch <- ev
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}
