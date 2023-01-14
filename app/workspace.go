package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
)

func (a *Application) CreateWorkspace(ctx context.Context, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.CreateWorkspaceAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	ws, err := a.NewWorkspace(ctx, opts)
	if err != nil {
		a.Error(err, "constructing workspace", "name", opts.Name, "subject", subject)
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

func (a *Application) UpdateWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspace(ctx, otf.UpdateWorkspaceAction, spec)
	if err != nil {
		return nil, err
	}

	if err := opts.Valid(); err != nil {
		a.Error(err, "updating workspace", "subject", subject)
		return nil, err
	}

	// retain ref to existing name so a name change can be detected
	var name string
	updated, err := a.db.UpdateWorkspace(ctx, spec, func(ws *otf.Workspace) error {
		name = ws.Name()
		return ws.UpdateWithOptions(ctx, opts)
	})
	if err != nil {
		a.Error(err, "updating workspace", append(spec.LogFields(), "subject", subject)...)
		return nil, err
	}

	if updated.Name() != name {
		a.Publish(otf.Event{Type: otf.EventWorkspaceRenamed, Payload: updated})
	}

	a.V(0).Info("updated workspace", append(spec.LogFields(), "subject", subject)...)

	return updated, nil
}

func (a *Application) ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*otf.Workspace, error) {
	return a.db.ListWorkspacesByWebhookID(ctx, id)
}

func (a *Application) ConnectWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.ConnectWorkspaceOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspace(ctx, otf.UpdateWorkspaceAction, spec)
	if err != nil {
		return nil, err
	}

	ws, err := a.Connect(ctx, spec, opts)
	if err != nil {
		a.Error(err, "connecting workspace", append(spec.LogFields(), "subject", subject, "repo", opts.Identifier)...)
		return nil, err
	}

	a.V(0).Info("connected workspace repo", append(spec.LogFields(), "subject", subject, "repo", opts)...)

	return ws, nil
}

func (a *Application) UpdateWorkspaceRepo(ctx context.Context, spec otf.WorkspaceSpec, repo otf.WorkspaceRepo) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspace(ctx, otf.UpdateWorkspaceAction, spec)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.UpdateWorkspaceRepo(ctx, spec, repo)
	if err != nil {
		a.Error(err, "updating workspace repo connection", append(spec.LogFields(), "subject", subject, "repo", repo)...)
		return nil, err
	}

	a.V(0).Info("updated workspace repo connection", append(spec.LogFields(), "subject", subject, "repo", repo)...)

	return ws, nil
}

func (a *Application) DisconnectWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspace(ctx, otf.UpdateWorkspaceAction, spec)
	if err != nil {
		return nil, err
	}

	ws, err := a.Disconnect(ctx, spec)
	if err != nil {
		a.Error(err, "disconnecting workspace", append(spec.LogFields(), "subject", subject)...)
		return nil, err
	}

	a.V(0).Info("disconnected workspace", append(spec.LogFields(), "subject", subject)...)

	return ws, nil
}

func (a *Application) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	if opts.OrganizationName == nil {
		// subject needs perms on site to list workspaces across site
		_, err := a.CanAccessSite(ctx, otf.ListWorkspacesAction)
		if err != nil {
			return nil, err
		}
	} else {
		// check if subject has perms to list workspaces in organization
		_, err := a.CanAccessOrganization(ctx, otf.ListWorkspacesAction, *opts.OrganizationName)
		if err == otf.ErrAccessNotPermitted {
			// user does not have org-wide perms; fallback to listing workspaces
			// for which they have workspace-level perms.
			subject, err := otf.SubjectFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if user, ok := subject.(*otf.User); ok {
				return a.db.ListWorkspacesByUserID(ctx, user.ID(), *opts.OrganizationName, opts.ListOptions)
			}
		} else if err != nil {
			return nil, err
		}
	}

	return a.db.ListWorkspaces(ctx, opts)
}

func (a *Application) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspace(ctx, otf.GetWorkspaceAction, spec)
	if err != nil {
		return nil, err
	}

	if err := spec.Valid(); err != nil {
		a.Error(err, "retrieving workspace", "subject", subject)
		return nil, err
	}

	ws, err := a.db.GetWorkspace(ctx, spec)
	if err != nil {
		a.Error(err, "retrieving workspace", append(spec.LogFields(), "subject", subject)...)
		return nil, err
	}

	a.V(2).Info("retrieved workspace", append(spec.LogFields(), "subject", subject)...)

	return ws, nil
}

func (a *Application) DeleteWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspace(ctx, otf.DeleteWorkspaceAction, spec)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.GetWorkspace(ctx, spec)
	if err != nil {
		return nil, err
	}

	if err := a.db.DeleteWorkspace(ctx, spec); err != nil {
		a.Error(err, "deleting workspace", "id", ws.ID(), "name", ws.Name(), "subject", subject)
		return nil, err
	}

	a.Publish(otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws})

	a.V(0).Info("deleted workspace", "id", ws.ID(), "name", ws.Name(), "subject", subject)

	return ws, nil
}

func (a *Application) LockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspace(ctx, otf.LockWorkspaceAction, spec)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.LockWorkspace(ctx, spec, opts)
	if err != nil {
		a.Error(err, "locking workspace", append(spec.LogFields(), "subject", subject)...)
		return nil, err
	}
	a.V(1).Info("locked workspace", append(spec.LogFields(), "subject", subject)...)

	a.Publish(otf.Event{Type: otf.EventWorkspaceLocked, Payload: ws})

	return ws, nil
}

func (a *Application) UnlockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	subject, err := a.CanAccessWorkspace(ctx, otf.UnlockWorkspaceAction, spec)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.UnlockWorkspace(ctx, spec, opts)
	if err != nil {
		a.Error(err, "unlocking workspace", append(spec.LogFields(), "subject", subject)...)
		return nil, err
	}
	a.V(1).Info("unlocked workspace", append(spec.LogFields(), "subject", subject)...)

	a.Publish(otf.Event{Type: otf.EventWorkspaceUnlocked, Payload: ws})

	return ws, nil
}

// SetCurrentRun sets the current run for the workspace
func (a *Application) SetCurrentRun(ctx context.Context, workspaceID, runID string) error {
	return a.db.SetCurrentRun(ctx, workspaceID, runID)
}
