package app

import (
	"context"

	"github.com/leg100/otf"
)

// Watch provides authenticated access to a stream of events.
//
// NOTE: only events for workspaces and workspace related resources such as runs
// are watched.
func (a *Application) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	var err error
	if opts.WorkspaceID != nil {
		// caller must have workspace-level read permissions
		_, err = a.CanAccessWorkspaceByID(ctx, otf.WatchAction, *opts.WorkspaceID)
	} else if opts.Organization != nil {
		// caller must have organization-level read permissions
		_, err = a.CanAccessOrganization(ctx, otf.WatchAction, *opts.Organization)
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

				// watch only items that are either:
				// (a) workspaces
				// (b) resources that belong to a workspace (e.g. a run)
				if ws, ok := ev.Payload.(*otf.Workspace); ok {
					// apply workspace filter
					if opts.WorkspaceID != nil {
						if ws.Name() != *opts.WorkspaceID {
							continue
						}
					}
					// apply organization filter
					if opts.Organization != nil {
						if ws.Organization() != *opts.Organization {
							continue
						}
					}
				} else if res, ok := ev.Payload.(interface{ WorkspaceID() string }); ok {
					// apply workspace filter
					if opts.WorkspaceID != nil {
						if res.WorkspaceID() != *opts.WorkspaceID {
							continue
						}
					}
					// apply organization filter
					if opts.Organization != nil {
						// fetch workspace first in order to get organization
						// name
						ws, err := a.GetWorkspace(ctx, res.WorkspaceID())
						if err != nil {
							a.Error(err, "retrieving workspace for watch event")
							continue
						}
						if ws.Organization() != *opts.Organization {
							continue
						}
					}
				} else {
					// skip all other events
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
