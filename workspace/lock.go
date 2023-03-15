package workspace

import "github.com/leg100/otf"

var (
	EventLocked   otf.EventType = "workspace_locked"
	EventUnlocked otf.EventType = "workspace_unlocked"
)

type (
	// Lock is a workspace lock, which blocks runs from running and prevents state from being
	// uploaded.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#locking
	Lock struct {
		LockedState // nil means unlocked
	}

	// LockedState is the workspace lock in a locked state, revealing who/what has
	// locked it and whether it can be locked/unlocked.
	LockedState interface {
		// CanLock checks whether it can be replaced with the given locked state
		CanLock(lock LockedState) error
		// CanUnlock checks whether subject is permitted to transfer it into the,
		// unlocked state, forceably or not.
		CanUnlock(lock LockedState, force bool) error
	}
)

// Locked determines whether lock is locked.
func (l *Lock) Locked() bool {
	return l.LockedState != nil
}

// Lock transfers a workspace into the given locked state
func (l *Lock) Lock(state LockedState) error {
	if !l.Locked() {
		// anything can lock an unlocked lock
		l.LockedState = state
		return nil
	}
	if err := l.LockedState.CanLock(state); err != nil {
		return err
	}
	l.LockedState = state
	return nil
}

// Unlock the lock.
func (l *Lock) Unlock(state LockedState, force bool) error {
	if !l.Locked() {
		return otf.ErrWorkspaceAlreadyUnlocked
	}
	if err := l.LockedState.CanUnlock(state, force); err != nil {
		return err
	}
	l.LockedState = nil
	return nil
}
