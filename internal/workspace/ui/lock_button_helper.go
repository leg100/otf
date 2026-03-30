package ui

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

type workspaceLockInfo struct {
	Text     string // button text
	Tooltip  string // button tooltip
	Disabled bool   // button greyed out or not
	Message  string // message accompanying button
	Action   string // form URL
}

// lockButtonHelper helps the UI determine the button to display for
// locking/unlocking the workspace.
func (h *Handlers) lockButtonHelper(
	ctx context.Context,
	ws *workspace.Workspace,
	currentUser *user.User,
) (workspaceLockInfo, error) {
	var btn workspaceLockInfo

	if ws.Locked() {
		btn.Text = "Unlock"
		btn.Action = paths.UnlockWorkspace(ws.ID)
		// A user needs at least the unlock permission
		if !h.Authorizer.CanAccess(ctx, authz.UnlockWorkspaceAction, ws.ID) {
			btn.Tooltip = "insufficient permissions"
			btn.Disabled = true
			return btn, nil
		}
		// Report who/what has locked the workspace.
		btn.Message = fmt.Sprintf("locked by: %s", ws.Lock)
		// also show message as button tooltip
		btn.Tooltip = btn.Message
		// A user can unlock their own lock
		if ws.Lock == currentUser.Username {
			return btn, nil
		}
		// User is going to need the force unlock permission
		if h.Authorizer.CanAccess(ctx, authz.ForceUnlockWorkspaceAction, ws.ID) {
			btn.Text = "Force unlock"
			btn.Action = paths.ForceUnlockWorkspace(ws.ID)
			return btn, nil
		}
		// User cannot unlock
		btn.Disabled = true
		return btn, nil
	} else {
		btn.Text = "Lock"
		btn.Action = paths.LockWorkspace(ws.ID)
		// User needs at least the lock permission
		if !h.Authorizer.CanAccess(ctx, authz.LockWorkspaceAction, ws.ID) {
			btn.Disabled = true
			btn.Tooltip = "insufficient permissions"
		}
		return btn, nil
	}
}
