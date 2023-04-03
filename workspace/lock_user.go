package workspace

import "github.com/leg100/otf"

// UserLock is a lock held by a user
type UserLock struct {
	ID, Username string
}

func (l UserLock) String() string { return l.Username }

func (l UserLock) CanLock(lock LockedState) error {
	// nothing can replace a user lock; it can only be unlocked
	return otf.ErrWorkspaceAlreadyLocked
}

func (l UserLock) CanUnlock(state LockedState, force bool) error {
	// only a user lock can unlock a user lock
	user, ok := state.(UserLock)
	if !ok {
		return otf.ErrWorkspaceUnlockDenied
	}

	// any user lock can forceably unlock a user lock
	if force {
		return nil
	}

	// otherwise only the same user lock can unlock a user lock
	if l == user {
		return nil
	}
	return otf.ErrWorkspaceLockedByDifferentUser
}
