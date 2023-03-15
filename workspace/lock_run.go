package workspace

import (
	"github.com/leg100/otf"
)

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
	return otf.ErrWorkspaceAlreadyLocked
}

func (RunLock) CanUnlock(lock LockedState, force bool) error {
	// a user lock is only allowed to unlock a run lock by force
	if _, ok := lock.(UserLock); ok {
		if force {
			return nil
		}
		return otf.ErrWorkspaceLockedByDifferentUser
	}
	// anyone/anything else is allowed to unlock a run lock
	return nil
}
