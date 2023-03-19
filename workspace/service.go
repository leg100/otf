package workspace

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/vcsprovider"
)

type (
	WorkspaceService   = Service
	VCSProviderService vcsprovider.Service

	Service interface {
		UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateWorkspaceOptions) (*Workspace, error)
		GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
		GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error)
		GetWorkspaceJSONAPI(ctx context.Context, workspaceID string, r *http.Request) (*jsonapi.Workspace, error)
		ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
		// ListWorkspacesByWebhookID retrieves workspaces by webhook ID.
		//
		// TODO: rename to ListConnectedWorkspaces
		ListWorkspacesByRepoID(ctx context.Context, repoID uuid.UUID) ([]*Workspace, error)
		//UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateWorkspaceOptions) (Workspace, error)
		DeleteWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)

		SetCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error)

		create(ctx context.Context, opts CreateWorkspaceOptions) (*Workspace, error)
		getRun(ctx context.Context, runID string) (run, error)
		delete(ctx context.Context, workspaceID string) (*Workspace, error)

		connect(ctx context.Context, workspaceID string, opts ConnectWorkspaceOptions) (*repo.Connection, error)
		disconnect(ctx context.Context, workspaceID string) error

		LockService
		permissionsService
	}

	service struct {
		logr.Logger
		otf.Publisher

		site           otf.Authorizer
		organization   otf.Authorizer
		otf.Authorizer // workspace authorizer

		db   *pgdb
		repo repo.RepoService

		*jsonapiMarshaler

		api *api
		web *webHandlers
	}

	Options struct {
		otf.DB
		*pubsub.Hub
		otf.Renderer
		organization.OrganizationService
		repo.RepoService
		auth.TeamService
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:    opts.Logger,
		Publisher: opts.Hub,
		repo:      opts.RepoService,
	}

	svc.organization = &organization.Authorizer{opts.Logger}
	svc.site = &otf.SiteAuthorizer{opts.Logger}

	svc.jsonapiMarshaler = &jsonapiMarshaler{
		Service:            opts.OrganizationService,
		permissionsService: &svc,
	}

	svc.api = &api{
		jsonapiMarshaler: svc.jsonapiMarshaler,
		svc:              &svc,
	}
	svc.web = &webHandlers{
		Renderer:    opts.Renderer,
		TeamService: opts.TeamService,
		svc:         &svc,
	}

	// Must register table name and service with pubsub broker so that it knows
	// how to lookup workspaces in the DB.
	opts.Register("workspaces", &svc)

	return serviceWithDB(&svc, newdb(opts.DB))
}

// serviceWithDB is for wrapping the service's db inside a tx
func serviceWithDB(svc *service, db *pgdb) *service {
	svc.db = db
	svc.Authorizer = &authorizer{
		Logger: svc.Logger,
		db:     db,
	}
	// TODO: construct connector

	return svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *service) CreateWorkspace(ctx context.Context, opts CreateWorkspaceOptions) (*Workspace, error) {
	return s.create(ctx, opts)
}

// GetByID implements pubsub.Getter
func (s *service) GetByID(ctx context.Context, workspaceID string) (any, error) {
	return s.db.GetWorkspace(ctx, workspaceID)
}

func (s *service) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.GetWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := s.db.GetWorkspace(ctx, workspaceID)
	if err != nil {
		s.Error(err, "retrieving workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}

	s.V(2).Info("retrieved workspace", "subject", subject, "workspace", workspaceID)

	return ws, nil
}

func (s *service) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	ws, err := s.db.GetWorkspaceByName(ctx, organization, workspace)
	if err != nil {
		s.Error(err, "retrieving workspace", "organization", organization, "workspace", workspace)
		return nil, err
	}

	subject, err := s.CanAccess(ctx, rbac.GetWorkspaceAction, ws.ID)
	if err != nil {
		return nil, err
	}

	s.V(2).Info("retrieved workspace", "subject", subject, "organization", organization, "workspace", workspace)

	return ws, nil
}

func (s *service) GetWorkspaceJSONAPI(ctx context.Context, workspaceID string, r *http.Request) (*jsonapi.Workspace, error) {
	ws, err := s.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return s.jsonapiMarshaler.toWorkspace(ws, r)
}

func (s *service) ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error) {
	if opts.Organization == nil {
		// subject needs perms on site to list workspaces across site
		_, err := s.site.CanAccess(ctx, rbac.ListWorkspacesAction, "")
		if err != nil {
			return nil, err
		}
	} else {
		// check if subject has perms to list workspaces in organization
		_, err := s.organization.CanAccess(ctx, rbac.ListWorkspacesAction, *opts.Organization)
		if err == otf.ErrAccessNotPermitted {
			// user does not have org-wide perms; fallback to listing workspaces
			// for which they have workspace-level perms.
			subject, err := otf.SubjectFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if user, ok := subject.(*auth.User); ok {
				return s.db.ListWorkspacesByUserID(ctx, user.ID, *opts.Organization, opts.ListOptions)
			}
		} else if err != nil {
			return nil, err
		}
	}

	return s.db.ListWorkspaces(ctx, opts)
}

func (s *service) ListWorkspacesByRepoID(ctx context.Context, repoID uuid.UUID) ([]*Workspace, error) {
	return s.db.ListWorkspacesByWebhookID(ctx, repoID)
}

func (s *service) UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateWorkspaceOptions) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	// retain ref to existing name so a name change can be detected
	var name string
	updated, err := s.db.UpdateWorkspace(ctx, workspaceID, func(ws *Workspace) error {
		name = ws.Name
		return ws.Update(opts)
	})
	if err != nil {
		s.Error(err, "updating workspace", "workspace", workspaceID, "subject", subject)
		return nil, err
	}

	if updated.Name != name {
		s.Publish(otf.Event{Type: otf.EventWorkspaceRenamed, Payload: updated})
	}

	s.V(0).Info("updated workspace", "workspace", workspaceID, "subject", subject)

	return updated, nil
}

func (s *service) DeleteWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	return nil, nil
}

func (s *service) create(ctx context.Context, opts CreateWorkspaceOptions) (*Workspace, error) {
	ws, err := NewWorkspace(opts)
	if err != nil {
		s.Error(err, "constructing workspace")
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.CreateWorkspaceAction, ws.Organization)
	if err != nil {
		return nil, err
	}

	err = s.tx(ctx, func(tx *service) error {
		if err := tx.db.CreateWorkspace(ctx, ws); err != nil {
			return err
		}
		// If needed, connect the VCS repository.
		if opts.ConnectWorkspaceOptions != nil {
			conn, err := tx.connect(ctx, ws.ID, *opts.ConnectWorkspaceOptions)
			if err != nil {
				return err
			}
			ws.Connection = conn
		}
		return nil
	})
	if err != nil {
		s.Error(err, "creating workspace", "id", ws.ID, "name", ws.Name, "organization", ws.Organization, "subject", subject)
		return nil, err
	}

	s.V(0).Info("created workspace", "id", ws.ID, "name", ws.Name, "organization", ws.Organization, "subject", subject)

	s.Publish(otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws})

	return ws, nil
}

func (s *service) connect(ctx context.Context, workspaceID string, opts ConnectWorkspaceOptions) (*repo.Connection, error) {
	subject, err := s.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	conn, err := s.repo.Connect(ctx, repo.ConnectOptions{
		ConnectionType: repo.WorkspaceConnection,
		ResourceID:     workspaceID,
		VCSProviderID:  opts.VCSProviderID,
		RepoPath:       opts.RepoPath,
	})
	if err != nil {
		s.Error(err, "connecting workspace", "workspace", workspaceID, "subject", subject, "repo", opts.RepoPath)
		return nil, err
	}

	s.V(0).Info("connected workspace repo", "workspace", workspaceID, "subject", subject, "repo", opts)

	return conn, nil
}

func (s *service) disconnect(ctx context.Context, workspaceID string) error {
	subject, err := s.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return err
	}

	err = s.repo.Disconnect(ctx, repo.DisconnectOptions{
		ConnectionType: repo.WorkspaceConnection,
		ResourceID:     workspaceID,
	})
	// ignore warnings; the repo is still disconnected successfully
	if err != nil && !errors.Is(err, otf.ErrWarning) {
		s.Error(err, "disconnecting workspace", "workspace", workspaceID, "subject", subject)
		return err
	}

	s.V(0).Info("disconnected workspace", "workspace", workspaceID, "subject", subject)

	return nil
}

// getRun retrieves a workspace run.
func (s *service) getRun(ctx context.Context, runID string) (run, error) {
	result, err := s.db.getRun(ctx, runID)
	if err != nil {
		s.Error(err, "retrieving workspace run", "run", runID)
		return run{}, err
	}

	s.V(2).Info("retrieved workspace run", "run", runID)

	return result, nil
}

func (s *service) delete(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.DeleteWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := s.db.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// disconnect repo before deleting
	if ws.Connection != nil {
		err = s.disconnect(ctx, ws.ID)
		// ignore warnings; the repo is still disconnected successfully
		if err != nil && !errors.Is(err, otf.ErrWarning) {
			return nil, err
		}
	}

	if err := s.db.DeleteWorkspace(ctx, ws.ID); err != nil {
		s.Error(err, "deleting workspace", "id", ws.ID, "name", ws.Name, "subject", subject)
		return nil, err
	}

	s.Publish(otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws})

	s.V(0).Info("deleted workspace", "id", ws.ID, "name", ws.Name, "subject", subject)

	return ws, nil
}

// SetCurrentRun sets the current run for the workspace
func (s *service) SetCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error) {
	return s.db.SetCurrentRun(ctx, workspaceID, runID)
}

// tx returns a service in a callback, with its database calls wrapped within a transaction
func (s *service) tx(ctx context.Context, txFunc func(*service) error) error {
	return s.db.Tx(ctx, func(db otf.DB) error {
		return txFunc(serviceWithDB(s, newdb(db)))
	})
}
