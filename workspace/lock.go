package workspace

import (
	"errors"

	"github.com/leg100/otf"
)

var (
	ErrWorkspaceAlreadyLocked         = errors.New("workspace already locked")
	ErrWorkspaceLockedByDifferentUser = errors.New("workspace locked by different user")
	ErrWorkspaceAlreadyUnlocked       = errors.New("workspace already unlocked")
	ErrWorkspaceUnlockDenied          = errors.New("unauthorized to unlock workspace")
	ErrWorkspaceInvalidLock           = errors.New("invalid workspace lock")

	EventWorkspaceLocked   otf.EventType = "workspace_locked"
	EventWorkspaceUnlocked otf.EventType = "workspace_unlocked"
)

// Unlocked is an unlocked workspace lock
type Unlocked struct {
	// zero identity because an unlocked workspace lock state has no identity
	otf.Identity
}

// CanLock always returns true
func (u *Unlocked) CanLock(otf.Identity) error {
	return nil
}

// CanUnlock always returns error
func (u *Unlocked) CanUnlock(otf.Identity, bool) error {
	return ErrWorkspaceAlreadyUnlocked
}
