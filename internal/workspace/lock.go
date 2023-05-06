package workspace

import (
	internal "github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/rbac"
)

const (
	UserLock LockKind = iota
	RunLock
)

var (
	EventLocked   internal.EventType = "workspace_locked"
	EventUnlocked internal.EventType = "workspace_unlocked"
)

type (
	// Lock is a workspace lock, which blocks runs from running and prevents state from being
	// uploaded.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#locking
	lock struct {
		id       string // ID of entity holding lock
		LockKind        // kind of entity holding lock
	}

	// kind of entity holding a lock
	LockKind int

	LockButton struct {
		State    string // locked or unlocked
		Text     string // button text
		Tooltip  string // button tooltip
		Disabled bool   // button greyed out or not
		Action   string // form URL
	}
)

// Locked determines whether workspace is locked.
func (ws *Workspace) Locked() bool {
	// a nil receiver means the lock is unlocked
	return ws.lock != nil
}

// Lock the workspace
func (ws *Workspace) Lock(id string, kind LockKind) error {
	if ws.lock == nil {
		ws.lock = &lock{
			id:       id,
			LockKind: kind,
		}
		return nil
	}
	// a run can replace another run holding a lock
	if kind == RunLock {
		ws.lock.id = id
		return nil
	}
	return internal.ErrWorkspaceAlreadyLocked
}

// Unlock the workspace.
func (ws *Workspace) Unlock(id string, kind LockKind, force bool) error {
	if ws.lock == nil {
		return internal.ErrWorkspaceAlreadyUnlocked
	}
	if force {
		ws.lock = nil
		return nil
	}
	// user can unlock their own lock
	if ws.LockKind == UserLock && kind == UserLock && ws.lock.id == id {
		ws.lock = nil
		return nil
	}
	// run can unlock its own lock
	if ws.LockKind == RunLock && kind == RunLock && ws.lock.id == id {
		ws.lock = nil
		return nil
	}
	// otherwise assume user is trying to unlock a lock held by a different
	// user (we don't assume a run because the scheduler should never attempt to
	// unlock a workspace using anything other than the original run).
	return internal.ErrWorkspaceLockedByDifferentUser
}

// lockButtonHelper helps the UI determine the button to display for
// locking/unlocking the workspace.
func lockButtonHelper(ws *Workspace, policy internal.WorkspacePolicy, user internal.Subject) LockButton {
	var btn LockButton

	if ws.Locked() {
		btn.State = "locked"
		btn.Text = "Unlock"
		btn.Action = paths.UnlockWorkspace(ws.ID)
		// A user needs at least the unlock permission
		if !user.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy) {
			btn.Tooltip = "insufficient permissions"
			btn.Disabled = true
			return btn
		}
		// Determine tooltip to show
		switch ws.LockKind {
		case UserLock:
			btn.Tooltip = "locked by user: " + ws.lock.id
		case RunLock:
			btn.Tooltip = "locked by run: " + ws.lock.id
		default:
			btn.Tooltip = "locked by unknown entity: " + ws.lock.id
		}
		// A user can unlock their own lock
		if ws.LockKind == UserLock && ws.lock.id == user.String() {
			return btn
		}
		// User is going to need the force unlock permission
		if user.CanAccessWorkspace(rbac.ForceUnlockWorkspaceAction, policy) {
			btn.Text = "Force unlock"
			btn.Action = paths.ForceUnlockWorkspace(ws.ID)
			return btn
		}
		// User cannot unlock
		btn.Disabled = true
		return btn
	} else {
		btn.State = "unlocked"
		btn.Text = "Lock"
		btn.Action = paths.LockWorkspace(ws.ID)
		// User needs at least the lock permission
		if !user.CanAccessWorkspace(rbac.LockWorkspaceAction, policy) {
			btn.Disabled = true
			btn.Tooltip = "insufficient permissions"
		}
		return btn
	}
}
