package workspace

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type application interface {
	create(ctx context.Context, opts CreateWorkspaceOptions) (*Workspace, error)
	get(ctx context.Context, workspaceID string) (*Workspace, error)
	getByName(ctx context.Context, organization, workspace string) (*Workspace, error)
	list(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
	listByWebhook(ctx context.Context, id uuid.UUID) ([]*Workspace, error)
	update(ctx context.Context, workspaceID string, opts UpdateWorkspaceOptions) (*Workspace, error)
	delete(ctx context.Context, workspaceID string) (*Workspace, error)

	connect(ctx context.Context, workspaceID string, opts connectOptions) error
	disconnect(ctx context.Context, workspaceID string) (*Workspace, error)

	lockApp
	permissionsApp
}

type app struct {
	otf.Authorizer
	otf.PubSubService
	logr.Logger

	connector *Connector
	db        *pgdb
}

func (a *app) update(ctx context.Context, workspaceID string, opts UpdateWorkspaceOptions) (*Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	// retain ref to existing name so a name change can be detected
	var name string
	updated, err := a.db.UpdateWorkspace(ctx, workspaceID, func(ws *Workspace) error {
		name = ws.Name()
		return ws.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating workspace", "workspace", workspaceID, "subject", subject)
		return nil, err
	}

	if updated.Name() != name {
		a.Publish(otf.Event{Type: otf.EventWorkspaceRenamed, Payload: updated})
	}

	a.V(0).Info("updated workspace", "workspace", workspaceID, "subject", subject)

	return updated, nil
}

func (a *app) listByWebhook(ctx context.Context, id uuid.UUID) ([]*Workspace, error) {
	return a.db.ListWorkspacesByWebhookID(ctx, id)
}

func (a *app) connect(ctx context.Context, workspaceID string, opts connectOptions) error {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return err
	}

	err = a.connector.connect(ctx, workspaceID, opts)
	if err != nil {
		a.Error(err, "connecting workspace", "workspace", workspaceID, "subject", subject, "repo", opts.Identifier)
		return err
	}

	a.V(0).Info("connected workspace repo", "workspace", workspaceID, "subject", subject, "repo", opts)

	return nil
}

func (a *app) UpdateWorkspaceRepo(ctx context.Context, workspaceID string, repo WorkspaceRepo) (*Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.UpdateWorkspaceRepo(ctx, workspaceID, repo)
	if err != nil {
		a.Error(err, "updating workspace repo connection", "workspace", workspaceID, "subject", subject, "repo", repo)
		return nil, err
	}

	a.V(0).Info("updated workspace repo connection", "workspace", workspaceID, "subject", subject, "repo", repo)

	return ws, nil
}

func (a *app) disconnect(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err = a.connector.disconnect(ctx, ws)
	if err != nil {
		a.Error(err, "disconnecting workspace", "workspace", workspaceID, "subject", subject)
		return nil, err
	}

	a.V(0).Info("disconnected workspace", "workspace", workspaceID, "subject", subject)

	return ws, nil
}

func (a *app) list(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error) {
	if opts.Organization == nil {
		// subject needs perms on site to list workspaces across site
		_, err := a.CanAccessSite(ctx, rbac.ListWorkspacesAction)
		if err != nil {
			return nil, err
		}
	} else {
		// check if subject has perms to list workspaces in organization
		_, err := a.CanAccessOrganization(ctx, rbac.ListWorkspacesAction, *opts.Organization)
		if err == otf.ErrAccessNotPermitted {
			// user does not have org-wide perms; fallback to listing workspaces
			// for which they have workspace-level perms.
			subject, err := otf.SubjectFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if user, ok := subject.(otf.User); ok {
				return a.db.ListWorkspacesByUserID(ctx, user.ID(), *opts.Organization, opts.ListOptions)
			}
		} else if err != nil {
			return nil, err
		}
	}

	return a.db.ListWorkspaces(ctx, opts)
}

func (a *app) get(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.GetWorkspaceAction, workspaceID)
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

func (a *app) getByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	ws, err := a.db.GetWorkspaceByName(ctx, organization, workspace)
	if err != nil {
		a.Error(err, "retrieving workspace", "organization", organization, "workspace", workspace)
		return nil, err
	}

	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.GetWorkspaceAction, ws.id)
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved workspace", "subject", subject, "organization", organization, "workspace", workspace)

	return ws, nil
}

func (a *app) delete(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.DeleteWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	if err := a.db.DeleteWorkspace(ctx, ws.ID()); err != nil {
		a.Error(err, "deleting workspace", "id", ws.ID(), "name", ws.Name(), "subject", subject)
		return nil, err
	}

	a.Publish(otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws})

	a.V(0).Info("deleted workspace", "id", ws.ID(), "name", ws.Name(), "subject", subject)

	return ws, nil
}

// SetCurrentRun sets the current run for the workspace
func (a *app) setCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error) {
	return a.db.SetCurrentRun(ctx, workspaceID, runID)
}

func (a *app) create(ctx context.Context, opts CreateWorkspaceOptions) (*Workspace, error) {
	ws, err := NewWorkspace(opts)
	if err != nil {
		a.Error(err, "constructing workspace")
		return nil, err
	}

	subject, err := a.CanAccessOrganization(ctx, rbac.CreateWorkspaceAction, ws.Organization())
	if err != nil {
		return nil, err
	}

	if err := a.db.CreateWorkspace(ctx, ws); err != nil {
		a.Error(err, "creating workspace", "id", ws.ID(), "name", ws.Name(), "organization", ws.Organization(), "subject", subject)
		return nil, err
	}

	a.V(0).Info("created workspace", "id", ws.ID(), "name", ws.Name(), "organization", ws.Organization(), "subject", subject)

	a.Publish(otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws})

	return ws, nil
}
