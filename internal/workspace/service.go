package workspace

import (
	"context"

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
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/state"
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
		ListWorkspaces(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error)
		// ListWorkspacesByWebhookID retrieves workspaces by webhook ID.
		//
		// TODO: rename to ListConnectedWorkspaces
		ListWorkspacesByRepoID(ctx context.Context, repoID uuid.UUID) ([]*Workspace, error)
		DeleteWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)

		SetCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error)

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
		*sql.DB
		*pubsub.Broker
		html.Renderer
		organization.OrganizationService
		vcsprovider.VCSProviderService
		state.StateService
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
		StateService:       opts.StateService,
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

	err = s.db.Tx(ctx, func(ctx context.Context, _ pggen.Querier) error {
		if err := s.db.create(ctx, ws); err != nil {
			return err
		}
		// Optionally connect workspace to repo.
		if ws.Connection != nil {
			if err := s.connect(ctx, ws.ID, ws.Connection); err != nil {
				return err
			}
		}
		// Optionally create tags.
		if len(opts.Tags) > 0 {
			added, err := s.addTags(ctx, ws, opts.Tags)
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

func (s *service) ListWorkspaces(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
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
				return s.db.listByUsername(ctx, user.Username, *opts.Organization, opts.PageOptions)
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

	// update the workspace and optionally connect/disconnect to/from vcs repo.
	var updated *Workspace
	err = s.db.Tx(ctx, func(ctx context.Context, _ pggen.Querier) error {
		var connect *bool
		updated, err = s.db.update(ctx, workspaceID, func(ws *Workspace) (err error) {
			connect, err = ws.Update(opts)
			return err
		})
		if err != nil {
			return err
		}
		if connect != nil {
			if *connect {
				if err := s.connect(ctx, workspaceID, updated.Connection); err != nil {
					return err
				}
			} else {
				if err := s.disconnect(ctx, workspaceID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "updating workspace", "workspace", workspaceID, "subject", subject)
		return nil, err
	}

	s.V(0).Info("updated workspace", "workspace", workspaceID, "subject", subject)

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
		if err := s.disconnect(ctx, ws.ID); err != nil {
			return nil, err
		}
	}

	if err := s.db.delete(ctx, ws.ID); err != nil {
		s.Error(err, "deleting workspace", "id", ws.ID, "name", ws.Name, "subject", subject)
		return nil, err
	}

	s.V(0).Info("deleted workspace", "id", ws.ID, "name", ws.Name, "subject", subject)

	return ws, nil
}

// connect connects the workspace to a repo.
func (s *service) connect(ctx context.Context, workspaceID string, connection *Connection) error {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = s.repo.Connect(ctx, repo.ConnectOptions{
		ConnectionType: repo.WorkspaceConnection,
		ResourceID:     workspaceID,
		VCSProviderID:  connection.VCSProviderID,
		RepoPath:       connection.Repo,
	})
	if err != nil {
		s.Error(err, "connecting workspace", "workspace", workspaceID, "subject", subject, "repo", connection.Repo)
		return err
	}
	s.V(0).Info("connected workspace repo", "workspace", workspaceID, "subject", subject, "repo", connection.Repo)

	return nil
}

func (s *service) disconnect(ctx context.Context, workspaceID string) error {
	subject, err := internal.SubjectFromContext(ctx)
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
