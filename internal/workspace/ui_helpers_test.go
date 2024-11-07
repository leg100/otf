package workspace

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_LockButtonHelper(t *testing.T) {
	janitorTestID = resource.NewID(resource.UserKind)

	tests := []struct {
		name string
		lock *lock
		user *fakeUser
		want LockButton
	}{
		{
			"unlocked state",
			&lock{},
			&fakeUser{canLock: true},
			LockButton{
				State:  "unlocked",
				Text:   "Lock",
				Action: "/app/workspaces/ws-123/lock",
			},
		},
		{
			"insufficient permissions to lock",
			&lock{},
			&fakeUser{},
			LockButton{
				State:    "unlocked",
				Text:     "Lock",
				Tooltip:  "insufficient permissions",
				Action:   "/app/workspaces/ws-123/lock",
				Disabled: true,
			},
		},
		{
			"insufficient permissions to unlock",
			&lock{ID: &janitorTestID},
			&fakeUser{},
			LockButton{
				State:    "locked",
				Text:     "Unlock",
				Tooltip:  "insufficient permissions",
				Action:   "/app/workspaces/ws-123/unlock",
				Disabled: true,
			},
		},
		{
			"user can unlock their own lock",
			&lock{ID: &janitorTestID},
			&fakeUser{id: "janitor", canUnlock: true},
			LockButton{
				State:   "locked",
				Text:    "Unlock",
				Message: "locked by: janitor",
				Tooltip: "locked by: janitor",
				Action:  "/app/workspaces/ws-123/unlock",
			},
		},
		{
			"can unlock lock held by a different user",
			&lock{ID: &janitorTestID},
			&fakeUser{id: "burglar", canUnlock: true},
			LockButton{
				State:    "locked",
				Text:     "Unlock",
				Action:   "/app/workspaces/ws-123/unlock",
				Message:  "locked by: janitor",
				Tooltip:  "locked by: janitor",
				Disabled: true,
			},
		},
		{
			"user can force unlock",
			&lock{ID: &janitorTestID},
			&fakeUser{id: "headmaster", canUnlock: true, canForceUnlock: true},
			LockButton{
				State:   "locked",
				Text:    "Force unlock",
				Action:  "/app/workspaces/ws-123/force-unlock",
				Message: "locked by: janitor",
				Tooltip: "locked by: janitor",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := &uiHelpers{
				service: &fakeUIHelpersService{"janitor"},
			}
			workspaceID := resource.ParseID("ws-123")
			got, err := helpers.lockButtonHelper(context.Background(), workspaceID, tt.lock, authz.WorkspacePolicy{}, tt.user)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeUIHelpersService struct {
	username string
}

func (f *fakeUIHelpersService) GetUser(context.Context, user.UserSpec) (*user.User, error) {
	return &user.User{Username: f.username}, nil
}

type fakeUser struct {
	id                                 string
	canUnlock, canForceUnlock, canLock bool

	authz.Subject
}

func (f *fakeUser) String() string { return f.id }

func (f *fakeUser) CanAccessWorkspace(action rbac.Action, _ authz.WorkspacePolicy) bool {
	switch action {
	case rbac.UnlockWorkspaceAction:
		return f.canUnlock
	case rbac.ForceUnlockWorkspaceAction:
		return f.canForceUnlock
	case rbac.LockWorkspaceAction:
		return f.canLock
	default:
		return false

	}
}
