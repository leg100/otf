package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
)

func (a *Application) CreateWorkspace(ctx context.Context, opts otf.CreateWorkspaceOptions) (*otf.Workspace, error) {
	ws, err := otf.NewWorkspace(opts)
	if err != nil {
		a.error(err, "constructing workspace")
		return nil, err
	}

	subject, err := a.CanAccessOrganization(ctx, otf.CreateWorkspaceAction, ws.Organization())
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

func (a *Application) UpdateWorkspace(ctx context.Context, workspaceID string, opts otf.UpdateWorkspaceOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	// retain ref to existing name so a name change can be detected
	var name string
	updated, err := a.db.UpdateWorkspace(ctx, workspaceID, func(ws *otf.Workspace) error {
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

func (a *Application) ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*otf.Workspace, error) {
	return a.db.ListWorkspacesByWebhookID(ctx, id)
}

func (a *Application) ConnectWorkspace(ctx context.Context, workspaceID string, opts otf.ConnectWorkspaceOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.Connect(ctx, workspaceID, opts)
	if err != nil {
		a.Error(err, "connecting workspace", "workspace", workspaceID, "subject", subject, "repo", opts.Identifier)
		return nil, err
	}

	a.V(0).Info("connected workspace repo", "workspace", workspaceID, "subject", subject, "repo", opts)

	return ws, nil
}

func (a *Application) UpdateWorkspaceRepo(ctx context.Context, workspaceID string, repo otf.WorkspaceRepo) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.UpdateWorkspaceAction, workspaceID)
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

func (a *Application) DisconnectWorkspace(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.Disconnect(ctx, workspaceID)
	if err != nil {
		a.Error(err, "disconnecting workspace", "workspace", workspaceID, "subject", subject)
		return nil, err
	}

	a.V(0).Info("disconnected workspace", "workspace", workspaceID, "subject", subject)

	return ws, nil
}

func (a *Application) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	if opts.Organization == nil {
		// subject needs perms on site to list workspaces across site
		_, err := a.CanAccessSite(ctx, otf.ListWorkspacesAction)
		if err != nil {
			return nil, err
		}
	} else {
		// check if subject has perms to list workspaces in organization
		_, err := a.CanAccessOrganization(ctx, otf.ListWorkspacesAction, *opts.Organization)
		if err == otf.ErrAccessNotPermitted {
			// user does not have org-wide perms; fallback to listing workspaces
			// for which they have workspace-level perms.
			subject, err := otf.SubjectFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if user, ok := subject.(*otf.User); ok {
				return a.db.ListWorkspacesByUserID(ctx, user.ID(), *opts.Organization, opts.ListOptions)
			}
		} else if err != nil {
			return nil, err
		}
	}

	return a.db.ListWorkspaces(ctx, opts)
}

func (a *Application) GetWorkspace(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.GetWorkspaceAction, workspaceID)
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

func (a *Application) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByName(ctx, otf.GetWorkspaceAction, organization, workspace)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.GetWorkspaceByName(ctx, organization, workspace)
	if err != nil {
		a.Error(err, "retrieving workspace", "subject", subject, "organization", organization, "workspace", workspace)
		return nil, err
	}

	a.V(2).Info("retrieved workspace", "subject", subject, "organization", organization, "workspace", workspace)

	return ws, nil
}

func (a *Application) DeleteWorkspace(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.DeleteWorkspaceAction, workspaceID)
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

func (a *Application) LockWorkspace(ctx context.Context, workspaceID string, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.LockWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.LockWorkspace(ctx, workspaceID, opts)
	if err != nil {
		a.Error(err, "locking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}
	a.V(1).Info("locked workspace", "subject", subject, "workspace", workspaceID)

	a.Publish(otf.Event{Type: otf.EventWorkspaceLocked, Payload: ws})

	return ws, nil
}

func (a *Application) UnlockWorkspace(ctx context.Context, workspaceID string, opts otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.LockWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.UnlockWorkspace(ctx, workspaceID, opts)
	if err != nil {
		a.Error(err, "unlocking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}
	a.V(1).Info("unlocked workspace", "subject", subject, "workspace", workspaceID)

	a.Publish(otf.Event{Type: otf.EventWorkspaceUnlocked, Payload: ws})

	return ws, nil
}

// SetCurrentRun sets the current run for the workspace
func (a *Application) SetCurrentRun(ctx context.Context, workspaceID, runID string) error {
	return a.db.SetCurrentRun(ctx, workspaceID, runID)
}
