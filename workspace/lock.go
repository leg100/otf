package workspace

import (
	"github.com/leg100/otf"
)

var (
	EventLocked   otf.EventType = "workspace_locked"
	EventUnlocked otf.EventType = "workspace_unlocked"
)

// Lock is a workspace lock, which blocks runs from running and prevents state from being
// uploaded.
//
// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#locking
type Lock struct {
	state LockedState // nil means unlocked
}

// Locked determines whether lock is locked.
func (l *Lock) Locked() bool {
	return l.state != nil
}

// Lock transfers a workspace into the given locked state
func (l *Lock) Lock(state LockedState) error {
	if l.state == nil {
		// anything can lock an unlocked lock
		return nil
	}
	if err := l.state.CanLock(state); err != nil {
		return err
	}
	l.state = state
	return nil
}

// Unlock the lock. The given identity and toggling force determines
// whether permission is granted and the operation succeeds.
func (l *Lock) Unlock(iden otf.Identity, force bool) error {
	if l.state == nil {
		return otf.ErrWorkspaceAlreadyUnlocked
	}
	if err := l.state.CanUnlock(iden, force); err != nil {
		return err
	}
	l.state = nil
	return nil
}

// LockedState is the workspace lock in a locked state, revealing who/what has
// locked it and whether it can be locked/unlocked.
type LockedState interface {
	// Who/what has locked the workspace
	otf.Identity

	// CanLock checks whether it can be replaced with the given locked state
	CanLock(lock LockedState) error
	// CanUnlock checks whether subject is permitted to transfer it into the,
	// unlocked state, forceably or not.
	CanUnlock(subject any, force bool) error
}

// RunLock is a workspace lock held by a run
type RunLock struct {
	id string
}

func (l RunLock) ID() string     { return l.id }
func (l RunLock) String() string { return l.id }

func (RunLock) CanLock(lock LockedState) error {
	// a run lock can only be replaced by another run lock
	if _, ok := lock.(RunLock); ok {
		return nil
	}
	return otf.ErrWorkspaceAlreadyLocked
}

func (RunLock) CanUnlock(subject any, force bool) error {
	// users are only allowed to unlock a run lock forceably
	if _, ok := subject.(otf.User); ok {
		if force {
			return nil
		}
		return otf.ErrWorkspaceLockedByDifferentUser
	}
	// anyone/anything else is allowed to unlock a run lock
	return nil
}

// UserLock is a lock held by a user
type UserLock struct {
	id, username string
}

func (l UserLock) ID() string     { return l.id }
func (l UserLock) String() string { return l.username }

func (l UserLock) CanLock(lock LockedState) error {
	// nothing can replace a user lock; it can only be unlocked
	return otf.ErrWorkspaceAlreadyLocked
}

func (l UserLock) CanUnlock(subject any, force bool) error {
	// only users can unlock a user lock
	user, ok := subject.(otf.User)
	if !ok {
		return otf.ErrWorkspaceUnlockDenied
	}

	if force {
		// assume caller has checked user has necessary perms to forceably
		// unlock
		return nil
	}

	// a user can only unlock their own lock
	if l.id == user.ID() {
		return nil
	}
	return otf.ErrWorkspaceLockedByDifferentUser
}
