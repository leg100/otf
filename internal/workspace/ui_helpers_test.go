package workspace

import (
	"context"
	"slices"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_LockButtonHelper(t *testing.T) {
	bobby := &user.User{ID: resource.NewTfeID(resource.UserKind), Username: "bobby"}
	annie := &user.User{ID: resource.NewTfeID(resource.UserKind), Username: "annie"}

	tests := []struct {
		name                   string
		lockedBy               *user.User
		currentUser            *user.User
		currentUserPermissions []authz.Action
		want                   LockButton
	}{
		{
			"unlocked state",
			nil,
			nil,
			[]authz.Action{authz.LockWorkspaceAction},
			LockButton{
				State:  "unlocked",
				Text:   "Lock",
				Action: "/app/workspaces/ws-123/lock",
			},
		},
		{
			"insufficient permissions to lock",
			nil,
			nil,
			[]authz.Action{},
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
			bobby,
			nil,
			[]authz.Action{},
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
			bobby,
			bobby,
			[]authz.Action{authz.UnlockWorkspaceAction},
			LockButton{
				State:   "locked",
				Text:    "Unlock",
				Message: "locked by: bobby",
				Tooltip: "locked by: bobby",
				Action:  "/app/workspaces/ws-123/unlock",
			},
		},
		{
			"user without force-unlock permission cannot force-unlock lock held by different user",
			bobby,
			annie,
			[]authz.Action{authz.UnlockWorkspaceAction},
			LockButton{
				State:    "locked",
				Text:     "Unlock",
				Message:  "locked by: bobby",
				Tooltip:  "locked by: bobby",
				Disabled: true,
				Action:   "/app/workspaces/ws-123/unlock",
			},
		},
		{
			"user with force-unlock permission can force-unlock lock held by different user",
			bobby,
			annie,
			[]authz.Action{authz.UnlockWorkspaceAction, authz.ForceUnlockWorkspaceAction},
			LockButton{
				State:   "locked",
				Text:    "Force unlock",
				Action:  "/app/workspaces/ws-123/force-unlock",
				Message: "locked by: bobby",
				Tooltip: "locked by: bobby",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := &uiHelpers{
				service:    &fakeUIHelpersService{lockedBy: tt.lockedBy},
				authorizer: &fakeLockButtonAuthorizer{perms: tt.currentUserPermissions},
			}
			ws := &Workspace{ID: testutils.ParseID(t, "ws-123")}
			if tt.lockedBy != nil {
				ws.Lock = &tt.lockedBy.ID
			}
			got, err := helpers.lockButtonHelper(context.Background(), ws, tt.currentUser)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeUIHelpersService struct {
	lockedBy *user.User
}

func (f *fakeUIHelpersService) GetUser(context.Context, user.UserSpec) (*user.User, error) {
	return f.lockedBy, nil
}

type fakeLockButtonAuthorizer struct {
	perms []authz.Action
}

func (f *fakeLockButtonAuthorizer) CanAccess(ctx context.Context, action authz.Action, _ resource.ID) bool {
	return slices.Contains(f.perms, action)
}
