package workspace

import "github.com/leg100/otf"

// UserLock is a lock held by a user
type UserLock struct {
	id, username string
}

func (l UserLock) String() string { return l.username }

func (l UserLock) CanLock(lock otf.LockedState) error {
	// nothing can replace a user lock; it can only be unlocked
	return otf.ErrWorkspaceAlreadyLocked
}

func (l UserLock) CanUnlock(subject any, force bool) error {
	// only users can unlock a user lock
	user, ok := subject.(*otf.User)
	if !ok {
		return otf.ErrWorkspaceUnlockDenied
	}

	if force {
		// assume caller has checked user has necessary perms to forceably
		// unlock
		return nil
	}

	// a user can only unlock their own lock
	if l.id == user.ID {
		return nil
	}
	return otf.ErrWorkspaceLockedByDifferentUser
}
