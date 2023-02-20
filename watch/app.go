package watch

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type application interface {
	Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error)
}

type app struct {
	otf.Authorizer
	otf.PubSubService
	logr.Logger
}

// Watch provides authenticated access to a stream of events.
//
// NOTE: only events for workspaces and workspace related resources such as runs
// are watched.
func (a *app) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	var err error
	if opts.WorkspaceID != nil {
		// caller must have workspace-level read permissions
		_, err = a.CanAccessWorkspaceByID(ctx, rbac.WatchAction, *opts.WorkspaceID)
	} else if opts.Organization != nil {
		// caller must have organization-level read permissions
		_, err = a.CanAccessOrganization(ctx, rbac.WatchAction, *opts.Organization)
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

				// Event must either be for a workspace or a resource that
				// belongs to a workspace, or the event is skipped.
				res, ok := ev.Payload.(otf.WorkspaceResource)
				if !ok {
					continue
				}

				// apply workspace filter
				if opts.WorkspaceID != nil {
					if res.WorkspaceID() != *opts.WorkspaceID {
						continue
					}
				}
				// apply organization filter
				if opts.Organization != nil {
					if res.Organization() != *opts.Organization {
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
