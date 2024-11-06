package workspace

import (
	"errors"

	"github.com/leg100/otf/internal/resource"
)

type (
	// lock is a workspace lock, which blocks runs from running and prevents state from being
	// uploaded.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#locking
	lock struct {
		*resource.ID
	}
)

func (l *lock) String() string {
	if l.ID == nil {
		return ""
	}
	return l.ID.String()
}

func (l *lock) Locked() bool {
	return l.ID != nil
}

// Enlock locks the lock with the given enlocker.
func (l *lock) Enlock(enlocker resource.ID) error {
	switch enlocker.Kind {
	case resource.UserKind, resource.RunKind:
	default:
		return errors.New("workspace can only be locked by a user or a run")
	}
	if !l.Locked() {
		l.ID = &enlocker
		return nil
	}
	// a run can replace another run holding a lock
	if l.Kind == resource.RunKind && enlocker.Kind == resource.RunKind {
		l.ID = &enlocker
		return nil
	}
	return ErrWorkspaceAlreadyLocked
}

// Unlock the workspace.
func (l *lock) Unlock(unlocker resource.ID, force bool) error {
	switch unlocker.Kind {
	case resource.UserKind, resource.RunKind:
	default:
		return errors.New("workspace can only be unlocked by a user or a run")
	}
	if l.ID == nil {
		return ErrWorkspaceAlreadyUnlocked
	}
	if force {
		l.ID = nil
		return nil
	}
	// user/run can unlock its own lock
	if *l.ID == unlocker {
		l.ID = nil
		return nil
	}
	if l.Kind == resource.RunKind {
		return ErrWorkspaceLockedByRun
	}
	return ErrWorkspaceLockedByDifferentUser
}
