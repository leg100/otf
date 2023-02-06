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

	EventWorkspaceLocked   EventType = "workspace_locked"
	EventWorkspaceUnlocked EventType = "workspace_unlocked"
)

// WorkspaceLockState is the state a workspace lock is currently in (i.e.
// unlocked, run-locked, or user-locked)
type WorkspaceLockState interface {
	// CanLock checks whether it can be locked by subject
	CanLock(subject otf.Identity) error
	// CanUnlock checks whether it can be unlocked by subject
	CanUnlock(subject otf.Identity, force bool) error
	// A lock state has an identity, i.e. the name of the run or user that has
	// locked the workspace
	otf.Identity
}

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
