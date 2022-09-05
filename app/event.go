package app

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

// Watch provides authenticated access to a stream of events. The returned channel is
// unbuffered.
func (a *Application) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	allowed, err := canAccessWatch(ctx, a.Mapper, opts)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, otf.ErrAccessNotPermitted
	}

	// channel for relaying events from subscription to caller
	ch := make(chan otf.Event)

	sub := a.Subscribe(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case event, ok := <-sub:
				if !ok {
					close(ch)
					return
				}
				ch <- event
			}
		}
	}()
	return ch, nil
}

func canAccessWatch(ctx context.Context, mapper *inmem.Mapper, opts otf.WatchOptions) (bool, error) {
	if opts.OrganizationName == nil && opts.WorkspaceID == nil && opts.WorkspaceName == nil {
		// Caller requesting access to *every* event
		return otf.CanAccess(ctx, nil), nil
	}
	spec := otf.WorkspaceSpec{
		OrganizationName: opts.OrganizationName,
		Name:             opts.WorkspaceName,
		ID:               opts.WorkspaceID,
	}
	if err := spec.Valid(); err != nil {
		return false, err
	}
	return mapper.CanAccessWorkspace(ctx, spec), nil
}
