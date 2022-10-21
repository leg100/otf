package app

import (
	"context"

	"github.com/leg100/otf"
)

// WorkspaceResource is a resource that belongs to a workspace and organization, or is a
// workspace
type WorkspaceResource interface {
	OrganizationName() string
	WorkspaceName() string
}

// Watch provides authenticated access to a stream of events.
//
// TODO: apply watch options
func (a *Application) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	var err error
	if opts.OrganizationName != nil && opts.WorkspaceName != nil {
		// caller must have workspace-level read permissions
		_, err = a.CanAccessWorkspace(ctx, otf.WatchAction, otf.WorkspaceSpec{
			Name:             opts.WorkspaceName,
			OrganizationName: opts.OrganizationName,
		})
	} else if opts.OrganizationName != nil {
		// caller must have organization-level read permissions
		_, err = a.CanAccessOrganization(ctx, otf.WatchAction, *opts.OrganizationName)
	}
	if err != nil {
		return nil, err
	}

	if opts.Name == nil {
		opts.Name = otf.String("watch-" + otf.GenerateRandomString(6))
	}
	sub, err := a.Subscribe(ctx, *opts.Name)
	if err != nil {
		return nil, err
	}

	ch := make(chan otf.Event)
	go func() {
		for {
			select {
			case ev, ok := <-sub:
				if !ok {
					close(ch)
					return
				}
				res, ok := ev.Payload.(WorkspaceResource)
				if !ok {
					// skip events that contain payloads that cannot be related
					// back to a workspace, including log updates which are
					// very noisy
					continue
				}
				// apply optional organization filter
				if opts.OrganizationName != nil {
					if res.OrganizationName() != *opts.OrganizationName {
						continue
					}
				}
				// apply optional workspace filter
				if opts.WorkspaceName != nil {
					if res.WorkspaceName() != *opts.WorkspaceName {
						continue
					}
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
// NOTE: unauthenticated. External access to this endpoint should only be made
// via signed URLs.
func (a *Application) WatchLogs(ctx context.Context, opts otf.WatchLogsOptions) (<-chan otf.Chunk, error) {
	name := "watch-logs-" + otf.GenerateRandomString(6)
	if opts.Name != nil {
		name = *opts.Name
	}

	ch := make(chan otf.Chunk)
	sub, err := a.Subscribe(ctx, name)
	if err != nil {
		return nil, err
	}

	go func() {
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
