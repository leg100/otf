package workspace

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
)

type service interface {
	create(ctx context.Context, opts otf.CreateWorkspaceOptions) (*otf.Workspace, error)
	get(ctx context.Context, workspaceID string) (*otf.Workspace, error)
	getByName(ctx context.Context, organization, workspace string) (*otf.Workspace, error)
	list(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error)
	listByWebhook(ctx context.Context, id uuid.UUID) ([]*otf.Workspace, error)
	update(ctx context.Context, workspaceID string, opts otf.UpdateWorkspaceOptions) (*otf.Workspace, error)
	delete(ctx context.Context, workspaceID string) (*otf.Workspace, error)

	connect(ctx context.Context, workspaceID string, opts otf.ConnectWorkspaceOptions) error
	disconnect(ctx context.Context, workspaceID string) error

	lockService
	permissionsService
}

type Service struct {
	logr.Logger
	otf.PubSubService

	site         otf.Authorizer
	organization otf.Authorizer
	*authorizer

	db   *pgdb
	repo otf.RepoService

	api *api
	web *web
}

func NewService(opts Options) *Service {
	svc := Service{
		Logger:        opts.Logger,
		PubSubService: opts.PubSubService,
		repo:          opts.RepoService,
		db:            newdb(opts.DB),
	}

	svc.organization = &organization.Authorizer{opts.Logger}
	svc.site = &otf.SiteAuthorizer{opts.Logger}

	svc.api = &api{
		svc:             &svc,
		tokenMiddleware: opts.TokenMiddleware,
	}
	svc.web = &web{
		Renderer:          opts.Renderer,
		svc:               &svc,
		sessionMiddleware: opts.SessionMiddleware,
	}

	return serviceWithDB(&svc, newdb(opts.DB))
}

func serviceWithDB(parent *Service, db *pgdb) *Service {
	child := *parent
	child.db = db
	child.authorizer = &authorizer{
		Logger: parent.Logger,
		db:     db,
	}
	// TODO: construct connector

	return &child
}

type Options struct {
	TokenMiddleware, SessionMiddleware mux.MiddlewareFunc

	otf.DB
	otf.PubSubService
	otf.Renderer
	otf.RepoService
	logr.Logger
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
	a.web.addHandlers(r)
}

func (s *Service) CreateWorkspace(ctx context.Context, opts otf.CreateWorkspaceOptions) (*otf.Workspace, error) {
	return s.create(ctx, opts)
}

func (s *Service) GetWorkspace(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	return nil, nil
}

func (s *Service) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*otf.Workspace, error) {
	return nil, nil
}

func (s *Service) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return nil, nil
}

func (s *Service) ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*otf.Workspace, error) {
	return nil, nil
}

func (s *Service) UpdateWorkspace(ctx context.Context, workspaceID string, opts otf.UpdateWorkspaceOptions) (*otf.Workspace, error) {
	return nil, nil
}

func (s *Service) DeleteWorkspace(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	return nil, nil
}

func (s *Service) LockWorkspace(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	return s.lock(ctx, workspaceID, nil)
}

func (s *Service) UnlockWorkspace(ctx context.Context, workspaceID string, force bool) (*otf.Workspace, error) {
	return s.unlock(ctx, workspaceID, force)
}

func (a *Service) create(ctx context.Context, opts otf.CreateWorkspaceOptions) (*otf.Workspace, error) {
	ws, err := otf.NewWorkspace(opts)
	if err != nil {
		a.Error(err, "constructing workspace")
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.CreateWorkspaceAction, ws.Organization)
	if err != nil {
		return nil, err
	}

	err = a.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.CreateWorkspace(ctx, ws); err != nil {
			return err
		}
		// If needed, connect the VCS repository.
		if repo := opts.Repo; repo != nil {
			return serviceWithDB(a, tx).connect(ctx, ws.ID, otf.ConnectWorkspaceOptions{
				ProviderID: repo.VCSProviderID,
				Identifier: repo.Identifier,
			})
		}
		return nil
	})
	if err != nil {
		a.Error(err, "creating workspace", "id", ws.ID, "name", ws.Name, "organization", ws.Organization, "subject", subject)
	}

	a.V(0).Info("created workspace", "id", ws.ID, "name", ws.Name, "organization", ws.Organization, "subject", subject)

	a.Publish(otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws})

	return ws, nil
}

func (a *Service) update(ctx context.Context, workspaceID string, opts otf.UpdateWorkspaceOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	// retain ref to existing name so a name change can be detected
	var name string
	updated, err := a.db.UpdateWorkspace(ctx, workspaceID, func(ws *otf.Workspace) error {
		name = ws.Name
		return ws.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating workspace", "workspace", workspaceID, "subject", subject)
		return nil, err
	}

	if updated.Name != name {
		a.Publish(otf.Event{Type: otf.EventWorkspaceRenamed, Payload: updated})
	}

	a.V(0).Info("updated workspace", "workspace", workspaceID, "subject", subject)

	return updated, nil
}

func (a *Service) listByWebhook(ctx context.Context, id uuid.UUID) ([]*otf.Workspace, error) {
	return a.db.ListWorkspacesByWebhookID(ctx, id)
}

func (a *Service) connect(ctx context.Context, workspaceID string, opts otf.ConnectWorkspaceOptions) error {
	subject, err := a.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return err
	}

	_, err = a.repo.Connect(ctx, otf.ConnectOptions{
		ConnectionType: otf.WorkspaceConnection,
		ResourceID:     workspaceID,
		VCSProviderID:  opts.ProviderID,
		Identifier:     opts.Identifier,
	})
	if err != nil {
		a.Error(err, "connecting workspace", "workspace", workspaceID, "subject", subject, "repo", opts.Identifier)
		return err
	}

	a.V(0).Info("connected workspace repo", "workspace", workspaceID, "subject", subject, "repo", opts)

	return nil
}

func (a *Service) disconnect(ctx context.Context, workspaceID string) error {
	subject, err := a.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return err
	}

	err = a.repo.Disconnect(ctx, otf.DisconnectOptions{
		ConnectionType: otf.WorkspaceConnection,
		ResourceID:     workspaceID,
	})
	// ignore warnings; the repo is still disconnected successfully
	if err != nil && !errors.Is(err, otf.ErrWarning) {
		a.Error(err, "disconnecting workspace", "workspace", workspaceID, "subject", subject)
		return err
	}

	a.V(0).Info("disconnected workspace", "workspace", workspaceID, "subject", subject)

	return nil
}

func (a *Service) list(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	if opts.Organization == nil {
		// subject needs perms on site to list workspaces across site
		_, err := a.site.CanAccess(ctx, rbac.ListWorkspacesAction, "")
		if err != nil {
			return nil, err
		}
	} else {
		// check if subject has perms to list workspaces in organization
		_, err := a.organization.CanAccess(ctx, rbac.ListWorkspacesAction, *opts.Organization)
		if err == otf.ErrAccessNotPermitted {
			// user does not have org-wide perms; fallback to listing workspaces
			// for which they have workspace-level perms.
			subject, err := otf.SubjectFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if user, ok := subject.(*otf.User); ok {
				return a.db.ListWorkspacesByUserID(ctx, user.ID, *opts.Organization, opts.ListOptions)
			}
		} else if err != nil {
			return nil, err
		}
	}

	return a.db.ListWorkspaces(ctx, opts)
}

func (a *Service) get(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	subject, err := a.CanAccess(ctx, rbac.GetWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.GetWorkspace(ctx, workspaceID)
	if err != nil {
		a.Error(err, "retrieving workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}

	a.V(2).Info("retrieved workspace", "subject", subject, "workspace", workspaceID)

	return ws, nil
}

func (a *Service) getByName(ctx context.Context, organization, workspace string) (*otf.Workspace, error) {
	ws, err := a.db.GetWorkspaceByName(ctx, organization, workspace)
	if err != nil {
		a.Error(err, "retrieving workspace", "organization", organization, "workspace", workspace)
		return nil, err
	}

	subject, err := a.CanAccess(ctx, rbac.GetWorkspaceAction, ws.ID)
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved workspace", "subject", subject, "organization", organization, "workspace", workspace)

	return ws, nil
}

func (a *Service) delete(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	subject, err := a.CanAccess(ctx, rbac.DeleteWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// disconnect repo before deleting
	if ws.Repo != nil {
		err = a.disconnect(ctx, ws.ID)
		// ignore warnings; the repo is still disconnected successfully
		if err != nil && !errors.Is(err, otf.ErrWarning) {
			return nil, err
		}
	}

	if err := a.db.DeleteWorkspace(ctx, ws.ID); err != nil {
		a.Error(err, "deleting workspace", "id", ws.ID, "name", ws.Name, "subject", subject)
		return nil, err
	}

	a.Publish(otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws})

	a.V(0).Info("deleted workspace", "id", ws.ID, "name", ws.Name, "subject", subject)

	return ws, nil
}

// SetCurrentRun sets the current run for the workspace
func (a *Service) setCurrentRun(ctx context.Context, workspaceID, runID string) (*otf.Workspace, error) {
	return a.db.SetCurrentRun(ctx, workspaceID, runID)
}

// tx returns a service in a callback, with its database calls wrapped within a transaction
func (a *Service) tx(ctx context.Context, txFunc func(*Service) error) error {
	return a.db.Tx(ctx, func(db otf.DB) error {
		return txFunc(serviceWithDB(a, newdb(db)))
	})
}
