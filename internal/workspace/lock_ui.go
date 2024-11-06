package workspace

import (
	"fmt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/user"
)

type LockButton struct {
	State    string // locked or unlocked
	Text     string // button text
	Tooltip  string // button tooltip
	Disabled bool   // button greyed out or not
	Message  string // message accompanying button
	Action   string // form URL
}

// lockButtonHelper helps the UI determine the button to display for
// locking/unlocking the workspace.
func lockButtonHelper(ws *Workspace, policy authz.WorkspacePolicy, user *user.User) LockButton {
	var btn LockButton

	if ws.Lock.Locked() {
		btn.State = "locked"
		btn.Text = "Unlock"
		btn.Action = paths.UnlockWorkspace(ws.ID.String())
		// A user needs at least the unlock permission
		if !user.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy) {
			btn.Tooltip = "insufficient permissions"
			btn.Disabled = true
			return btn
		}
		btn.Message = fmt.Sprintf("locked by: %s", ws.Lock)
		// also show message as button tooltip
		btn.Tooltip = btn.Message
		// A user can unlock their own lock
		if *ws.Lock.ID == user.ID {
			return btn
		}
		// User is going to need the force unlock permission
		if user.CanAccessWorkspace(rbac.ForceUnlockWorkspaceAction, policy) {
			btn.Text = "Force unlock"
			btn.Action = paths.ForceUnlockWorkspace(ws.ID.String())
			return btn
		}
		// User cannot unlock
		btn.Disabled = true
		return btn
	} else {
		btn.State = "unlocked"
		btn.Text = "Lock"
		btn.Action = paths.LockWorkspace(ws.ID.String())
		// User needs at least the lock permission
		if !user.CanAccessWorkspace(rbac.LockWorkspaceAction, policy) {
			btn.Disabled = true
			btn.Tooltip = "insufficient permissions"
		}
		return btn
	}
}
