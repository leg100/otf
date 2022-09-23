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
				if !ok {
					close(ch)
					return
				}
				res, ok := ev.Payload.(OrganizationResource)
				if !ok {
					// skip events that contain payloads that cannot be related
					// back to an organization, including log updates, which are
					// very noisy
					continue
				}
				if !otf.CanAccess(ctx, otf.String(res.OrganizationName())) {
					// skip events caller is not entitled to
					continue
				}

				ch <- ev
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}()
	return ch, nil
}

// WatchLogs provides a subscription to phase logs.
//
// NOTE: unauthenticated.
func (a *Application) WatchLogs(ctx context.Context, opts otf.WatchLogsOptions) (<-chan otf.Chunk, error) {
	ch := make(chan otf.Chunk)
	go func() {
		sub := a.Subscribe(ctx)
		for {
			select {
			case ev, ok := <-sub:
				if !ok {
					close(ch)
					return
				}
				chunk, ok := ev.Payload.(otf.PersistedChunk)
				if !ok {
					// skip non-log events
					continue
				}

				ch <- chunk.Chunk
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}()
	return ch, nil
}
