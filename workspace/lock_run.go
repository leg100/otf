package workspace

import "github.com/leg100/otf"

// RunLock is a workspace lock held by a run
type RunLock struct {
	id string
}

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
	if _, ok := subject.(*otf.User); ok {
		if force {
			return nil
		}
		return otf.ErrWorkspaceLockedByDifferentUser
	}
	// anyone/anything else is allowed to unlock a run lock
	return nil
}
