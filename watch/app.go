package watch

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/r3labs/sse/v2"
)

type Application struct {
	otf.Authorizer
	otf.PubSubService
	logr.Logger
	otf.WorkspaceService

	*api
	*htmlApp
}

func NewApplication(opts ApplicationOptions) *Application {
	app := &Application{
		Authorizer:       opts.Authorizer,
		Logger:           opts.Logger,
		PubSubService:    opts.PubSubService,
		WorkspaceService: opts.WorkspaceService,
	}

	// Create and configure SSE server
	srv := sse.New()
	// we don't use last-event-item functionality so turn it off
	srv.AutoReplay = false
	// encode payloads into base64 otherwise the JSON string payloads corrupt
	// the SSE protocol
	srv.EncodeBase64 = true

	app.api = &api{
		Application:  app,
		eventsServer: srv,
	}
	app.htmlApp = &htmlApp{
		Application: app,
		Renderer:    opts.Renderer,
		Server:      srv,
	}
	return app
}

type ApplicationOptions struct {
	otf.Authorizer
	otf.PubSubService
	otf.Renderer
	otf.WorkspaceService
	logr.Logger
}

// Watch provides authenticated access to a stream of events.
//
// NOTE: only events for workspaces and workspace related resources such as runs
// are watched.
func (a *Application) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
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

				// watch only items that are either:
				// (a) workspaces
				// (b) resources that belong to a workspace (e.g. a run)
				if ws, ok := ev.Payload.(otf.Workspace); ok {
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
