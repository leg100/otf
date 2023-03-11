package watch

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
	"github.com/r3labs/sse/v2"
)

type (
	service interface {
		Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error)
	}

	Service struct {
		logr.Logger
		otf.PubSubService

		organization otf.Authorizer
		workspace    otf.Authorizer

		api *api
		web *web
	}

	Options struct {
		WorkspaceAuthorizer otf.Authorizer

		otf.PubSubService
		otf.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:        opts.Logger,
		PubSubService: opts.PubSubService,
	}

	svc.organization = &organization.Authorizer{opts.Logger}
	svc.workspace = opts.WorkspaceAuthorizer

	// Create and configure SSE server
	srv := newSSEServer()

	svc.api = &api{
		svc:          &svc,
		eventsServer: srv,
	}
	svc.web = &web{
		Renderer: opts.Renderer,
		svc:      &svc,
		Server:   srv,
	}
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

// Watch provides authenticated access to a stream of events.
//
// NOTE: only events for runs and workspaces are currently watched.
func (s *Service) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	var err error
	if opts.WorkspaceID != nil {
		// caller must have workspace-level read permissions
		_, err = s.workspace.CanAccess(ctx, rbac.WatchAction, *opts.WorkspaceID)
	} else if opts.Organization != nil {
		// caller must have organization-level read permissions
		_, err = s.organization.CanAccess(ctx, rbac.WatchAction, *opts.Organization)
	}
	if err != nil {
		return nil, err
	}

	if opts.Name == nil {
		opts.Name = otf.String("watch-" + otf.GenerateRandomString(6))
	}
	sub, err := s.Subscribe(ctx, *opts.Name)
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

				var organization, workspaceID string
				switch payload := ev.Payload.(type) {
				case *run.Run:
					organization = payload.Organization
					workspaceID = payload.WorkspaceID
				case *otf.Workspace:
					organization = payload.Organization
					workspaceID = payload.ID
				default:
					continue // skip anything other than a run or a workspace
				}

				// apply workspace filter
				if opts.WorkspaceID != nil {
					if workspaceID != *opts.WorkspaceID {
						continue
					}
				}
				// apply organization filter
				if opts.Organization != nil {
					if organization != *opts.Organization {
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

func newSSEServer() *sse.Server {
	// Create and configure SSE server
	srv := sse.New()
	// we don't use last-event-item functionality so turn it off
	srv.AutoReplay = false
	// encode payloads into base64 otherwise the JSON string payloads corrupt
	// the SSE protocol
	srv.EncodeBase64 = true

	return srv
}
