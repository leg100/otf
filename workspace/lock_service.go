package workspace

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type lockService interface {
	lock(ctx context.Context, workspaceID string, runID *string) (*otf.Workspace, error)
	unlock(ctx context.Context, workspaceID string, force bool) (*otf.Workspace, error)
}

// lock the workspace. A workspace can only be locked on behalf of a run or a
// user. If the former then runID must be populated. Otherwise a user is
// extracted from the context.
func (svc *Service) lock(ctx context.Context, workspaceID string, runID *string) (*otf.Workspace, error) {
	subject, err := svc.CanAccess(ctx, rbac.LockWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	var state LockedState
	if runID != nil {
		state = RunLock{id: *runID}
	} else if user, ok := subject.(otf.User); ok {
		state = UserLock{id: user.ID, username: user.Username()}
	} else {
		svc.Error(otf.ErrWorkspaceUnlockDenied, "subject", subject, "workspace", workspaceID)
		return nil, otf.ErrWorkspaceUnlockDenied
	}

	ws, err := svc.db.toggleLock(ctx, workspaceID, func(ws *otf.Workspace) error {
		return ws.Lock.Lock(state)
	})
	if err != nil {
		svc.Error(err, "locking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}
	svc.V(1).Info("locked workspace", "subject", subject, "workspace", workspaceID)

	svc.Publish(otf.Event{Type: EventLocked, Payload: ws})

	return ws, nil
}

func (svc *Service) unlock(ctx context.Context, workspaceID string, force bool) (*otf.Workspace, error) {
	action := rbac.UnlockWorkspaceAction
	if force {
		action = rbac.ForceUnlockWorkspaceAction
	}

	subject, err := svc.CanAccess(ctx, action, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := svc.db.toggleLock(ctx, workspaceID, func(ws *otf.Workspace) error {
		return ws.Unlock(subject, force)
	})
	if err != nil {
		svc.Error(err, "unlocking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}
	svc.V(1).Info("unlocked workspace", "subject", subject, "workspace", workspaceID)

	svc.Publish(otf.Event{Type: EventUnlocked, Payload: ws})

	return ws, nil
}
