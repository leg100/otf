package workspace

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type Service struct {
	application

	api *api
	web *web
}

func NewService(opts Options) (*Service, error) {
	app := &app{
		Authorizer:    opts.Authorizer,
		PubSubService: opts.PubSubService,
		db:            newPGDB(opts.DB),
		Logger:        opts.Logger,
	}
	api := &api{
		app:             app,
		tokenMiddleware: opts.TokenMiddleware,
	}
	web := &web{
		Renderer: opts.Renderer,
		app:      app,
	}

	return &Service{
		application: app,
		api:         api,
		web:         web,
	}, nil
}

type Options struct {
	TokenMiddleware mux.MiddlewareFunc

	otf.Authorizer
	otf.DB
	otf.PubSubService
	otf.Renderer
	logr.Logger
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
	a.web.addHandlers(r)
}

func (s *Service) CreateWorkspace(ctx context.Context, opts CreateWorkspaceOptions) (*Workspace, error) {
	return s.create(ctx, opts)
}

func (s *Service) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	return nil, nil
}

func (s *Service) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	return nil, nil
}

func (s *Service) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*WorkspaceList, error) {
	return nil, nil
}

func (s *Service) ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*Workspace, error) {
	return nil, nil
}

func (s *Service) UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateWorkspaceOptions) (*Workspace, error) {
	return nil, nil
}

func (s *Service) DeleteWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	return nil, nil
}

func (s *Service) LockWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	return s.lock(ctx, workspaceID, nil)
}

func (s *Service) UnlockWorkspace(ctx context.Context, workspaceID string, force bool) (*Workspace, error) {
	return s.unlock(ctx, workspaceID, force)
}
