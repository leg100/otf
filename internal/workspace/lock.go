package workspace

import (
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
)

const (
	UserLock LockKind = iota
	RunLock
)

type (
	// Lock is a workspace Lock, which blocks runs from running and prevents state from being
	// uploaded.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#locking
	Lock struct {
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
		Message  string // message accompanying button
		Action   string // form URL
	}
)

// Locked determines whether workspace is locked.
func (ws *Workspace) Locked() bool {
	// a nil receiver means the lock is unlocked
	return ws.Lock != nil
}

// Enlock locks the workspace
func (ws *Workspace) Enlock(id string, kind LockKind) error {
	if ws.Lock == nil {
		ws.Lock = &Lock{
			id:       id,
			LockKind: kind,
		}
		return nil
	}
	// a run can replace another run holding a lock
	if kind == RunLock {
		ws.Lock.id = id
		return nil
	}
	return internal.ErrWorkspaceAlreadyLocked
}

// Unlock the workspace.
func (ws *Workspace) Unlock(id string, kind LockKind, force bool) error {
	if ws.Lock == nil {
		return internal.ErrWorkspaceAlreadyUnlocked
	}
	if force {
		ws.Lock = nil
		return nil
	}
	// user can unlock their own lock
	if ws.Lock.LockKind == UserLock && kind == UserLock && ws.Lock.id == id {
		ws.Lock = nil
		return nil
	}
	// run can unlock its own lock
	if ws.Lock.LockKind == RunLock && kind == RunLock && ws.Lock.id == id {
		ws.Lock = nil
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
		// Determine message to show
		switch ws.Lock.LockKind {
		case UserLock, RunLock:
			btn.Message = "locked by: " + ws.Lock.id
		default:
			btn.Message = "locked by unknown entity: " + ws.Lock.id
		}
		// also show message as button tooltip
		btn.Tooltip = btn.Message
		// A user can unlock their own lock
		if ws.Lock.LockKind == UserLock && ws.Lock.id == user.String() {
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
