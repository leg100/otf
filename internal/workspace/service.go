package workspace

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	Service struct {
		logr.Logger
		*authz.Authorizer

		db          *pgdb
		web         *webHandlers
		tfeapi      *tfe
		api         *api
		broker      *pubsub.Broker[*Workspace]
		connections *connections.Service

		beforeCreateHooks []func(context.Context, *Workspace) error
		afterCreateHooks  []func(context.Context, *Workspace) error
		beforeUpdateHooks []func(context.Context, *Workspace) error
	}

	Options struct {
		*sql.DB
		*sql.Listener
		*tfeapi.Responder
		*authz.Authorizer

		logr.Logger

		OrganizationService *organization.Service
		VCSProviderService  *vcsprovider.Service
		TeamService         *team.Service
		UserService         *user.Service
		ConnectionService   *connections.Service
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		Logger:      opts.Logger,
		Authorizer:  opts.Authorizer,
		db:          db,
		connections: opts.ConnectionService,
	}
	svc.web = newWebHandlers(&svc, opts)
	svc.tfeapi = &tfe{
		Service:    &svc,
		Responder:  opts.Responder,
		Authorizer: opts.Authorizer,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}
	svc.broker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"workspaces",
		func(ctx context.Context, id resource.ID, action sql.Action) (*Workspace, error) {
			if action == sql.DeleteAction {
				return &Workspace{ID: id}, nil
			}
			return db.get(ctx, id)
		},
	)
	// Fetch workspace when API calls request workspace be included in the
	// response
	opts.Responder.Register(tfeapi.IncludeWorkspace, svc.tfeapi.include)
	opts.Responder.Register(tfeapi.IncludeWorkspaces, svc.tfeapi.includeMany)
	// Instruct the authorizer to resolve workspace IDs to organization names.
	opts.Authorizer.RegisterOrganizationResolver(resource.WorkspaceKind, func(ctx context.Context, id resource.ID) (string, error) {
		ws, err := svc.db.get(ctx, id)
		if err != nil {
			return "", err
		}
		return ws.Organization, nil
	})
	// Provide the authorizer with the ability to retrieve workspace policies.
	opts.Authorizer.WorkspacePolicyGetter = &svc
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
	s.tfeapi.addHandlers(r)
	s.web.addTagHandlers(r)
	s.tfeapi.addTagHandlers(r)
	s.api.addHandlers(r)
}

func (s *Service) Watch(ctx context.Context) (<-chan pubsub.Event[*Workspace], func()) {
	return s.broker.Subscribe(ctx)
}

func (s *Service) Create(ctx context.Context, opts CreateOptions) (*Workspace, error) {
	ws, err := NewWorkspace(opts)
	if err != nil {
		s.Error(err, "constructing workspace")
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.CreateWorkspaceAction, &authz.AccessRequest{Organization: ws.Organization})
	if err != nil {
		return nil, err
	}

	err = s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		for _, hook := range s.beforeCreateHooks {
			if err := hook(ctx, ws); err != nil {
				return err
			}
		}
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
		for _, hook := range s.afterCreateHooks {
			if err := hook(ctx, ws); err != nil {
				return err
			}
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

func (s *Service) BeforeCreateWorkspace(hook func(context.Context, *Workspace) error) {
	s.beforeCreateHooks = append(s.beforeCreateHooks, hook)
}

func (s *Service) AfterCreateWorkspace(hook func(context.Context, *Workspace) error) {
	s.afterCreateHooks = append(s.afterCreateHooks, hook)
}

func (s *Service) Get(ctx context.Context, workspaceID resource.ID) (*Workspace, error) {
	subject, err := s.Authorize(ctx, authz.GetWorkspaceAction, &authz.AccessRequest{ID: &workspaceID})
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

func (s *Service) GetByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	ws, err := s.db.getByName(ctx, organization, workspace)
	if err != nil {
		s.Error(err, "retrieving workspace", "organization", organization, "workspace", workspace)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.GetWorkspaceAction, &authz.AccessRequest{ID: &ws.ID})
	if err != nil {
		return nil, err
	}

	s.V(9).Info("retrieved workspace", "subject", subject, "organization", organization, "workspace", workspace)

	return ws, nil
}

func (s *Service) List(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	if opts.Organization == nil {
		// subject needs perms on site to list workspaces across site
		_, err := s.Authorize(ctx, authz.ListWorkspacesAction, nil)
		if err != nil {
			return nil, err
		}
	} else {
		// check if subject has perms to list workspaces in organization
		_, err := s.Authorize(ctx, authz.ListWorkspacesAction, &authz.AccessRequest{Organization: *opts.Organization})
		if err == internal.ErrAccessNotPermitted {
			// user does not have org-wide perms; fallback to listing workspaces
			// for which they have workspace-level perms.
			subject, err := authz.SubjectFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if user, ok := subject.(*user.User); ok {
				return s.db.listByUsername(ctx, user.Username, *opts.Organization, opts.PageOptions)
			}
		} else if err != nil {
			return nil, err
		}
	}

	return s.db.list(ctx, opts)
}

func (s *Service) ListConnectedWorkspaces(ctx context.Context, vcsProviderID resource.ID, repoPath string) ([]*Workspace, error) {
	return s.db.listByConnection(ctx, vcsProviderID, repoPath)
}

func (s *Service) BeforeUpdateWorkspace(hook func(context.Context, *Workspace) error) {
	s.beforeUpdateHooks = append(s.beforeUpdateHooks, hook)
}

func (s *Service) Update(ctx context.Context, workspaceID resource.ID, opts UpdateOptions) (*Workspace, error) {
	subject, err := s.Authorize(ctx, authz.UpdateWorkspaceAction, &authz.AccessRequest{ID: &workspaceID})
	if err != nil {
		return nil, err
	}

	// update the workspace and optionally connect/disconnect to/from vcs repo.
	var updated *Workspace
	err = s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		var connect *bool
		updated, err = s.db.update(ctx, workspaceID, func(ctx context.Context, ws *Workspace) (err error) {
			for _, hook := range s.beforeUpdateHooks {
				if err := hook(ctx, ws); err != nil {
					return err
				}
			}
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

func (s *Service) Delete(ctx context.Context, workspaceID resource.ID) (*Workspace, error) {
	subject, err := s.Authorize(ctx, authz.DeleteWorkspaceAction, &authz.AccessRequest{ID: &workspaceID})
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

func (s *Service) SetPermission(ctx context.Context, workspaceID, teamID resource.ID, role authz.Role) error {
	subject, err := s.Authorize(ctx, authz.SetWorkspacePermissionAction, &authz.AccessRequest{ID: &workspaceID})
	if err != nil {
		return err
	}

	if err := s.db.SetWorkspacePermission(ctx, workspaceID, teamID, role); err != nil {
		s.Error(err, "setting workspace permission", "subject", subject, "workspace", workspaceID)
		return err
	}

	s.V(0).Info("set workspace permission", "team_id", teamID, "role", role, "subject", subject, "workspace", workspaceID)

	// TODO: publish event

	return nil
}

func (s *Service) UnsetPermission(ctx context.Context, workspaceID, teamID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.UnsetWorkspacePermissionAction, &authz.AccessRequest{ID: &workspaceID})
	if err != nil {
		s.Error(err, "unsetting workspace permission", "team_id", teamID, "subject", subject, "workspace", workspaceID)
		return err
	}

	s.V(0).Info("unset workspace permission", "team_id", teamID, "subject", subject, "workspace", workspaceID)
	// TODO: publish event
	return s.db.UnsetWorkspacePermission(ctx, workspaceID, teamID)
}

// GetWorkspacePolicy retrieves the authorization policy for a workspace.
//
// NOTE: there is no auth because it is used in the process of making an auth
// decision.
func (s *Service) GetWorkspacePolicy(ctx context.Context, workspaceID resource.ID) (authz.WorkspacePolicy, error) {
	return s.db.GetWorkspacePolicy(ctx, workspaceID)
}

// connect connects the workspace to a repo.
func (s *Service) connect(ctx context.Context, workspaceID resource.ID, connection *Connection) error {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = s.connections.Connect(ctx, connections.ConnectOptions{
		ResourceID:    workspaceID,
		VCSProviderID: connection.VCSProviderID,
		RepoPath:      connection.Repo,
	})
	if err != nil {
		s.Error(err, "connecting workspace", "workspace", workspaceID, "subject", subject, "repo", connection.Repo)
		return err
	}
	s.V(0).Info("connected workspace repo", "workspace", workspaceID, "subject", subject, "repo", connection.Repo)

	return nil
}

func (s *Service) disconnect(ctx context.Context, workspaceID resource.ID) error {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.connections.Disconnect(ctx, connections.DisconnectOptions{
		ResourceID: workspaceID,
	})
	if err != nil {
		s.Error(err, "disconnecting workspace", "workspace", workspaceID, "subject", subject)
		return err
	}

	s.V(0).Info("disconnected workspace", "workspace", workspaceID, "subject", subject)

	return nil
}

// SetLatestRun sets the latest run for the workspace
func (s *Service) SetLatestRun(ctx context.Context, workspaceID, runID resource.ID) (*Workspace, error) {
	return s.db.setLatestRun(ctx, workspaceID, runID)
}
