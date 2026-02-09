package ui

import (
	"context"
	"slices"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_LockButtonHelper(t *testing.T) {
	bobby := user.NewTestUser(t)
	annie := user.NewTestUser(t)

	tests := []struct {
		name                   string
		lockedBy               *user.User
		currentUser            *user.User
		currentUserPermissions []authz.Action
		want                   workspaceLockInfo
	}{
		{
			"unlocked state",
			nil,
			nil,
			[]authz.Action{authz.LockWorkspaceAction},
			workspaceLockInfo{
				Text:   "Lock",
				Action: "/app/workspaces/ws-123/lock",
			},
		},
		{
			"insufficient permissions to lock",
			nil,
			nil,
			[]authz.Action{},
			workspaceLockInfo{
				Text:     "Lock",
				Tooltip:  "insufficient permissions",
				Action:   "/app/workspaces/ws-123/lock",
				Disabled: true,
			},
		},
		{
			"insufficient permissions to unlock",
			bobby,
			nil,
			[]authz.Action{},
			workspaceLockInfo{
				Text:     "Unlock",
				Tooltip:  "insufficient permissions",
				Action:   "/app/workspaces/ws-123/unlock",
				Disabled: true,
			},
		},
		{
			"user can unlock their own lock",
			bobby,
			bobby,
			[]authz.Action{authz.UnlockWorkspaceAction},
			workspaceLockInfo{
				Text:    "Unlock",
				Message: "locked by: " + bobby.String(),
				Tooltip: "locked by: " + bobby.String(),
				Action:  "/app/workspaces/ws-123/unlock",
			},
		},
		{
			"user without force-unlock permission cannot force-unlock lock held by different user",
			bobby,
			annie,
			[]authz.Action{authz.UnlockWorkspaceAction},
			workspaceLockInfo{
				Text:     "Unlock",
				Message:  "locked by: " + bobby.String(),
				Tooltip:  "locked by: " + bobby.String(),
				Disabled: true,
				Action:   "/app/workspaces/ws-123/unlock",
			},
		},
		{
			"user with force-unlock permission can force-unlock lock held by different user",
			bobby,
			annie,
			[]authz.Action{authz.UnlockWorkspaceAction, authz.ForceUnlockWorkspaceAction},
			workspaceLockInfo{
				Text:    "Force unlock",
				Action:  "/app/workspaces/ws-123/force-unlock",
				Message: "locked by: " + bobby.String(),
				Tooltip: "locked by: " + bobby.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handlers{
				Authorizer: &fakeLockButtonAuthorizer{perms: tt.currentUserPermissions},
			}
			ws := &workspace.Workspace{ID: testutils.ParseID(t, "ws-123")}
			if tt.lockedBy != nil {
				ws.Lock = tt.lockedBy.Username
			}
			got, err := h.lockButtonHelper(context.Background(), ws, tt.currentUser)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeLockButtonAuthorizer struct {
	authz.Interface
	perms []authz.Action
}

func (f *fakeLockButtonAuthorizer) CanAccess(ctx context.Context, action authz.Action, _ resource.ID) bool {
	return slices.Contains(f.perms, action)
}
