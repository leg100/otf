package otf

import "errors"

var (
	ErrWorkspaceAlreadyLocked         = errors.New("workspace already locked")
	ErrWorkspaceLockedByDifferentUser = errors.New("workspace locked by different user")
	ErrWorkspaceAlreadyUnlocked       = errors.New("workspace already unlocked")
	ErrWorkspaceUnlockDenied          = errors.New("unauthorized to unlock workspace")
	ErrWorkspaceInvalidLock           = errors.New("invalid workspace lock")
)

type WorkspaceLock interface {
	// CanLock checks whether lock can be locked by requestor
	CanLock(requestor Identity) error
	// CanUnlock checks whether lock can be unlocked by requestor
	CanUnlock(requestor Identity, force bool) error
	// A lock is identifiable
	Identity
}

// Unlocked is an unlocked workspace lock
type Unlocked struct {
	// no identity
	Identity
}

// CanLock always returns true
func (u *Unlocked) CanLock(Identity) error {
	return nil
}

// CanUnlock always returns error
func (u *Unlocked) CanUnlock(Identity, bool) error {
	return ErrWorkspaceAlreadyUnlocked
}
