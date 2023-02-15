package workspace

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type lockApp interface {
	lock(ctx context.Context, workspaceID string, runID *string) (*Workspace, error)
	unlock(ctx context.Context, workspaceID string, force bool) (*Workspace, error)
}

// lock the workspace. A workspace can only be locked on behalf of a run or a
// user. If the former then runID must be populated. Otherwise a user is
// extracted from the context.
func (a *app) lock(ctx context.Context, workspaceID string, runID *string) (*Workspace, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.LockWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	var state LockedState
	if runID != nil {
		state = RunLock{id: *runID}
	} else if user, ok := subject.(otf.User); ok {
		state = UserLock{id: user.ID(), username: user.Username()}
	} else {
		a.Error(otf.ErrWorkspaceUnlockDenied, "subject", subject, "workspace", workspaceID)
		return nil, otf.ErrWorkspaceUnlockDenied
	}

	ws, err := a.db.toggleLock(ctx, workspaceID, func(ws *Workspace) error {
		return ws.Lock.Lock(state)
	})
	if err != nil {
		a.Error(err, "locking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}
	a.V(1).Info("locked workspace", "subject", subject, "workspace", workspaceID)

	a.Publish(otf.Event{Type: EventLocked, Payload: ws})

	return ws, nil
}

func (a *app) unlock(ctx context.Context, workspaceID string, force bool) (*Workspace, error) {
	action := rbac.UnlockWorkspaceAction
	if force {
		action = rbac.ForceUnlockWorkspaceAction
	}

	subject, err := a.CanAccessWorkspaceByID(ctx, action, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := a.db.toggleLock(ctx, workspaceID, func(ws *Workspace) error {
		return ws.Unlock(subject, force)
	})
	if err != nil {
		a.Error(err, "unlocking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}
	a.V(1).Info("unlocked workspace", "subject", subject, "workspace", workspaceID)

	a.Publish(otf.Event{Type: EventUnlocked, Payload: ws})

	return ws, nil
}
