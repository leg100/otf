package workspace

import (
	"github.com/leg100/otf"
)

var (
	EventLocked   otf.EventType = "workspace_locked"
	EventUnlocked otf.EventType = "workspace_unlocked"
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
	return otf.ErrWorkspaceAlreadyUnlocked
}
