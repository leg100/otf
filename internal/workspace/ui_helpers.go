package workspace

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	userpkg "github.com/leg100/otf/internal/user"
)

type uiHelpers struct {
	service uiHelpersService
}

type uiHelpersService interface {
	GetUser(context.Context, userpkg.UserSpec) (*userpkg.User, error)
}

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
func (h *uiHelpers) lockButtonHelper(
	ctx context.Context,
	workspaceID resource.ID,
	lock *lock,
	policy authz.WorkspacePolicy,
	user authz.Subject,
) (LockButton, error) {
	var btn LockButton

	if lock.Locked() {
		btn.State = "locked"
		btn.Text = "Unlock"
		btn.Action = paths.UnlockWorkspace(workspaceID.String())
		// A user needs at least the unlock permission
		if !user.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy) {
			btn.Tooltip = "insufficient permissions"
			btn.Disabled = true
			return btn, nil
		}
		// Report who/what has locked the workspace. If it is a user then fetch
		// their username.
		var lockedBy string
		if lock.Kind == resource.UserKind {
			lockUser, err := h.service.GetUser(ctx, userpkg.UserSpec{UserID: lock.ID})
			if err != nil {
				return LockButton{}, nil
			}
			lockedBy = lockUser.Username
		} else {
			lockedBy = lock.String()
		}
		btn.Message = fmt.Sprintf("locked by: %s", lockedBy)
		// also show message as button tooltip
		btn.Tooltip = btn.Message
		// A user can unlock their own lock
		if *lock.ID == user.GetID() {
			return btn, nil
		}
		// User is going to need the force unlock permission
		if user.CanAccessWorkspace(rbac.ForceUnlockWorkspaceAction, policy) {
			btn.Text = "Force unlock"
			btn.Action = paths.ForceUnlockWorkspace(workspaceID.String())
			return btn, nil
		}
		// User cannot unlock
		btn.Disabled = true
		return btn, nil
	} else {
		btn.State = "unlocked"
		btn.Text = "Lock"
		btn.Action = paths.LockWorkspace(workspaceID.String())
		// User needs at least the lock permission
		if !user.CanAccessWorkspace(rbac.LockWorkspaceAction, policy) {
			btn.Disabled = true
			btn.Tooltip = "insufficient permissions"
		}
		return btn, nil
	}
}
