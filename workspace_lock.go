package otf

var (
	EventLocked   EventType = "workspace_locked"
	EventUnlocked EventType = "workspace_unlocked"
)

// Lock is a workspace lock, which blocks runs from running and prevents state from being
// uploaded.
//
// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#locking
type Lock struct {
	LockedState // nil means unlocked
}

// Locked determines whether lock is locked.
func (l *Lock) Locked() bool {
	return l.LockedState != nil
}

// Lock transfers a workspace into the given locked state
func (l *Lock) Lock(state LockedState) error {
	if l.LockedState == nil {
		// anything can lock an unlocked lock
		return nil
	}
	if err := l.LockedState.CanLock(state); err != nil {
		return err
	}
	l.LockedState = state
	return nil
}

// Unlock the lock. The given identity and toggling force determines
// whether permission is granted and the operation succeeds.
func (l *Lock) Unlock(iden any, force bool) error {
	if l.LockedState == nil {
		return ErrWorkspaceAlreadyUnlocked
	}
	if err := l.LockedState.CanUnlock(iden, force); err != nil {
		return err
	}
	l.LockedState = nil
	return nil
}

// LockedState is the workspace lock in a locked state, revealing who/what has
// locked it and whether it can be locked/unlocked.
type LockedState interface {
	// CanLock checks whether it can be replaced with the given locked state
	CanLock(lock LockedState) error
	// CanUnlock checks whether subject is permitted to transfer it into the,
	// unlocked state, forceably or not.
	CanUnlock(subject any, force bool) error
}

// RunLock is a workspace lock held by a run
type RunLock struct {
	ID string
}

func (l RunLock) String() string { return l.ID }

func (RunLock) CanLock(lock LockedState) error {
	// a run lock can only be replaced by another run lock
	if _, ok := lock.(RunLock); ok {
		return nil
	}
	return ErrWorkspaceAlreadyLocked
}

func (RunLock) CanUnlock(subject any, force bool) error {
	// users are only allowed to unlock a run lock forceably
	if _, ok := subject.(User); ok {
		if force {
			return nil
		}
		return ErrWorkspaceLockedByDifferentUser
	}
	// anyone/anything else is allowed to unlock a run lock
	return nil
}

// UserLock is a lock held by a user
type UserLock struct {
	ID, Username string
}

func (l UserLock) String() string { return l.Username }

func (l UserLock) CanLock(lock LockedState) error {
	// nothing can replace a user lock; it can only be unlocked
	return ErrWorkspaceAlreadyLocked
}

func (l UserLock) CanUnlock(subject any, force bool) error {
	// only users can unlock a user lock
	user, ok := subject.(*User)
	if !ok {
		return ErrWorkspaceUnlockDenied
	}

	if force {
		// assume caller has checked user has necessary perms to forceably
		// unlock
		return nil
	}

	// a user can only unlock their own lock
	if l.ID == user.ID {
		return nil
	}
	return ErrWorkspaceLockedByDifferentUser
}
