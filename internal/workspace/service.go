package workspace

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/repo"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	WorkspaceService    = Service
	VCSProviderService  vcsprovider.Service
	OrganizationService organization.Service

	Service interface {
		CreateWorkspace(ctx context.Context, opts CreateOptions) (*Workspace, error)
		UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateOptions) (*Workspace, error)
		GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
		GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error)
		ListWorkspaces(ctx context.Context, opts ListOptions) (*WorkspaceList, error)
		// ListWorkspacesByWebhookID retrieves workspaces by webhook ID.
		//
		// TODO: rename to ListConnectedWorkspaces
		ListWorkspacesByRepoID(ctx context.Context, repoID uuid.UUID) ([]*Workspace, error)
		DeleteWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)

		SetCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error)

		connect(ctx context.Context, workspaceID string, opts ConnectOptions) (*repo.Connection, error)
		disconnect(ctx context.Context, workspaceID string) error

		LockService
		PermissionsService
		TagService
	}

	service struct {
		logr.Logger
		pubsub.Publisher

		site                internal.Authorizer
		organization        internal.Authorizer
		internal.Authorizer // workspace authorizer

		db   *pgdb
		repo repo.RepoService

		web *webHandlers
	}

	Options struct {
		internal.DB
		*pubsub.Broker
		html.Renderer
		organization.OrganizationService
		vcsprovider.VCSProviderService
		repo.RepoService
		auth.TeamService
		logr.Logger
	}
)

func NewService(opts Options) *service {
	db := &pgdb{opts.DB}
	svc := service{
		Logger:    opts.Logger,
		Publisher: opts.Broker,
		Authorizer: &authorizer{
			Logger: opts.Logger,
			db:     db,
		},
		db:           db,
		repo:         opts.RepoService,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &internal.SiteAuthorizer{Logger: opts.Logger},
	}
	svc.web = &webHandlers{
		Renderer:           opts.Renderer,
		TeamService:        opts.TeamService,
		VCSProviderService: opts.VCSProviderService,
		svc:                &svc,
	}
	// Register with broker so that it can relay workspace events
	opts.Register("workspaces", &svc)
	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
	s.web.addTagHandlers(r)
}

func (s *service) CreateWorkspace(ctx context.Context, opts CreateOptions) (*Workspace, error) {
	ws, err := NewWorkspace(opts)
	if err != nil {
		s.Error(err, "constructing workspace")
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.CreateWorkspaceAction, ws.Organization)
	if err != nil {
		return nil, err
	}

	err = s.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.create(ctx, ws); err != nil {
			return err
		}
		// Optionally connect workspace to repo. This is done within a
		// transaction, so if this fails so does the creation of the workspace.
		// And skip authorization, otherwise the authorizer looks up the
		// workspace and fails, because it does not exist yet because the transaction
		// has not yet completed.
		if opts.ConnectOptions != nil {
			conn, err := s.connectWithoutAuthz(ctx, ws.ID, tx, *opts.ConnectOptions)
			if err != nil {
				return err
			}
			ws.Connection = conn
		}
		// Optionally create tags within same transaction
		if len(opts.Tags) > 0 {
			added, err := addTags(ctx, tx, ws, opts.Tags)
			if err != nil {
				return err
			}
			ws.Tags = added
		}
		return nil
	})
	if err != nil {
		s.Error(err, "creating workspace", "id", ws.ID, "name", ws.Name, "organization", ws.Organization, "subject", subject)
		return nil, err
	}

	s.V(0).Info("created workspace", "id", ws.ID, "name", ws.Name, "organization", ws.Organization, "subject", subject)

	s.Publish(pubsub.NewCreatedEvent(ws))

	return ws, nil
}

// GetByID implements pubsub.Getter
func (s *service) GetByID(ctx context.Context, workspaceID string, action pubsub.DBAction) (any, error) {
	if action == pubsub.DeleteDBAction {
		return &Workspace{ID: workspaceID}, nil
	}
	return s.db.get(ctx, workspaceID)
}

func (s *service) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.GetWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := s.db.get(ctx, workspaceID)
	if err != nil {
		s.Error(err, "retrieving workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}

	s.V(9).Info("retrieved workspace", "subject", subject, "workspace", workspaceID)

	return ws, nil
}

func (s *service) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	ws, err := s.db.getByName(ctx, organization, workspace)
	if err != nil {
		s.Error(err, "retrieving workspace", "organization", organization, "workspace", workspace)
		return nil, err
	}

	subject, err := s.CanAccess(ctx, rbac.GetWorkspaceAction, ws.ID)
	if err != nil {
		return nil, err
	}

	s.V(9).Info("retrieved workspace", "subject", subject, "organization", organization, "workspace", workspace)

	return ws, nil
}

func (s *service) ListWorkspaces(ctx context.Context, opts ListOptions) (*WorkspaceList, error) {
	if opts.Organization == nil {
		// subject needs perms on site to list workspaces across site
		_, err := s.site.CanAccess(ctx, rbac.ListWorkspacesAction, "")
		if err != nil {
			return nil, err
		}
	} else {
		// check if subject has perms to list workspaces in organization
		_, err := s.organization.CanAccess(ctx, rbac.ListWorkspacesAction, *opts.Organization)
		if err == internal.ErrAccessNotPermitted {
			// user does not have org-wide perms; fallback to listing workspaces
			// for which they have workspace-level perms.
			subject, err := internal.SubjectFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if user, ok := subject.(*auth.User); ok {
				return s.db.listByUsername(ctx, user.Username, *opts.Organization, opts.ListOptions)
			}
		} else if err != nil {
			return nil, err
		}
	}

	return s.db.list(ctx, opts)
}

func (s *service) ListWorkspacesByRepoID(ctx context.Context, repoID uuid.UUID) ([]*Workspace, error) {
	return s.db.listByWebhookID(ctx, repoID)
}

func (s *service) UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateOptions) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	updated, err := s.db.update(ctx, workspaceID, func(ws *Workspace) error {
		return ws.Update(opts)
	})
	if err != nil {
		s.Error(err, "updating workspace", "workspace", workspaceID, "subject", subject)
		return nil, err
	}

	s.V(0).Info("updated workspace", "workspace", workspaceID, "subject", subject)

	s.Publish(pubsub.NewUpdatedEvent(updated))

	return updated, nil
}

func (s *service) DeleteWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.DeleteWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := s.db.get(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// disconnect repo before deleting
	if ws.Connection != nil {
		err = s.disconnect(ctx, ws.ID)
		// ignore warnings; the repo is still disconnected successfully
		if err != nil && !errors.Is(err, internal.ErrWarning) {
			return nil, err
		}
	}

	if err := s.db.delete(ctx, ws.ID); err != nil {
		s.Error(err, "deleting workspace", "id", ws.ID, "name", ws.Name, "subject", subject)
		return nil, err
	}

	s.Publish(pubsub.NewDeletedEvent(ws))

	s.V(0).Info("deleted workspace", "id", ws.ID, "name", ws.Name, "subject", subject)

	return ws, nil
}

// connect connects the workspace to a repo.
func (s *service) connect(ctx context.Context, workspaceID string, opts ConnectOptions) (*repo.Connection, error) {
	_, err := s.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	return s.connectWithoutAuthz(ctx, workspaceID, nil, opts)
}

// connectWithoutAuthz connects the workspace to a repo without checking whether
// subject has authorization.
func (s *service) connectWithoutAuthz(ctx context.Context, workspaceID string, tx internal.DB, opts ConnectOptions) (*repo.Connection, error) {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := s.repo.Connect(ctx, repo.ConnectOptions{
		ConnectionType: repo.WorkspaceConnection,
		ResourceID:     workspaceID,
		VCSProviderID:  opts.VCSProviderID,
		RepoPath:       opts.RepoPath,
		Tx:             tx,
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
	if err != nil {
		s.Error(err, "disconnecting workspace", "workspace", workspaceID, "subject", subject)
		return err
	}

	s.V(0).Info("disconnected workspace", "workspace", workspaceID, "subject", subject)

	return nil
}

// SetCurrentRun sets the current run for the workspace
func (s *service) SetCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error) {
	return s.db.setCurrentRun(ctx, workspaceID, runID)
}
