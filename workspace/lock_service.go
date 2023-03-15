package workspace

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/rbac"
)

type lockService interface {
	LockWorkspace(ctx context.Context, workspaceID string, runID *string) (*Workspace, error)
	UnlockWorkspace(ctx context.Context, workspaceID string, runID *string, force bool) (*Workspace, error)
}

// lock the workspace. A workspace can only be locked on behalf of a run or a
// user. If the former then runID must be populated. Otherwise a user is
// extracted from the context.
func (s *service) LockWorkspace(ctx context.Context, workspaceID string, runID *string) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.LockWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	state, err := GetLockedState(subject, runID)
	if err != nil {
		s.Error(err, "unlocking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}

	ws, err := s.db.toggleLock(ctx, workspaceID, func(lock *Lock) error {
		return lock.Lock(state)
	})
	if err != nil {
		s.Error(err, "locking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}
	s.V(1).Info("locked workspace", "subject", subject, "workspace", workspaceID)

	s.Publish(otf.Event{Type: EventLocked, Payload: ws})

	return ws, nil
}

// Unlock the workspace. A workspace can only be unlocked on behalf of a run or
// a user. If the former then runID must be non-nil; otherwise a user is
// extracted from the context.
func (s *service) UnlockWorkspace(ctx context.Context, workspaceID string, runID *string, force bool) (*Workspace, error) {
	action := rbac.UnlockWorkspaceAction
	if force {
		action = rbac.ForceUnlockWorkspaceAction
	}
	subject, err := s.CanAccess(ctx, action, workspaceID)
	if err != nil {
		return nil, err
	}

	state, err := GetLockedState(subject, runID)
	if err != nil {
		s.Error(err, "unlocking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}

	ws, err := s.db.toggleLock(ctx, workspaceID, func(lock *Lock) error {
		return lock.Unlock(state, force)
	})
	if err != nil {
		s.Error(err, "unlocking workspace", "subject", subject, "workspace", workspaceID)
		return nil, err
	}
	s.V(1).Info("unlocked workspace", "subject", subject, "workspace", workspaceID)

	s.Publish(otf.Event{Type: EventUnlocked, Payload: ws})

	return ws, nil
}

func GetLockedState(subject otf.Subject, runID *string) (LockedState, error) {
	var state LockedState
	if runID != nil {
		state = RunLock{ID: *runID}
	} else if user, ok := subject.(*auth.User); ok {
		state = UserLock{ID: user.ID, Username: user.Username}
	} else {
		return nil, otf.ErrWorkspaceInvalidLock
	}
	return state, nil
}
