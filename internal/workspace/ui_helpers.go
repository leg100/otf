package workspace

import (
	"context"
	"fmt"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html/components/paths"
	"github.com/leg100/otf/internal/resource"
	userpkg "github.com/leg100/otf/internal/user"
)

type uiHelpers struct {
	authorizer uiHelpersAuthorizer
}

type uiHelpersAuthorizer interface {
	CanAccess(context.Context, authz.Action, resource.ID) bool
}

type lockInfo struct {
	Text     string        // button text
	Tooltip  string        // button tooltip
	Disabled bool          // button greyed out or not
	Message  string        // message accompanying button
	Action   templ.SafeURL // form URL
}

// lockButtonHelper helps the UI determine the button to display for
// locking/unlocking the workspace.
func (h *uiHelpers) lockButtonHelper(
	ctx context.Context,
	ws *Workspace,
	user *userpkg.User,
) (lockInfo, error) {
	var btn lockInfo

	if ws.Locked() {
		btn.Text = "Unlock"
		btn.Action = paths.UnlockWorkspace(ws.ID)
		// A user needs at least the unlock permission
		if !h.authorizer.CanAccess(ctx, authz.UnlockWorkspaceAction, ws.ID) {
			btn.Tooltip = "insufficient permissions"
			btn.Disabled = true
			return btn, nil
		}
		// Report who/what has locked the workspace.
		btn.Message = fmt.Sprintf("locked by: %s", ws.Lock)
		// also show message as button tooltip
		btn.Tooltip = btn.Message
		// A user can unlock their own lock
		if ws.Lock == user.Username {
			return btn, nil
		}
		// User is going to need the force unlock permission
		if h.authorizer.CanAccess(ctx, authz.ForceUnlockWorkspaceAction, ws.ID) {
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
		if !h.authorizer.CanAccess(ctx, authz.LockWorkspaceAction, ws.ID) {
			btn.Disabled = true
			btn.Tooltip = "insufficient permissions"
		}
		return btn, nil
	}
}
