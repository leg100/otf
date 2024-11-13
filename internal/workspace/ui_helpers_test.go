package workspace

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_LockButtonHelper(t *testing.T) {
	wsID := testutils.ParseID(t, "ws-123")
	privilegedUser := &user.User{ID: resource.NewID(resource.UserKind), SiteAdmin: true}
	privilegedUser2 := &user.User{ID: resource.NewID(resource.UserKind), SiteAdmin: true}
	unprivilegedUser := &user.User{ID: resource.NewID(resource.UserKind), SiteAdmin: false}

	tests := []struct {
		name string
		ws   *Workspace
		user *user.User
		want LockButton
	}{
		{
			"unlocked state",
			&Workspace{ID: wsID},
			privilegedUser,
			LockButton{
				State:  "unlocked",
				Text:   "Lock",
				Action: "/app/workspaces/ws-123/lock",
			},
		},
		{
			"insufficient permissions to lock",
			&Workspace{ID: wsID},
			unprivilegedUser,
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
			&Workspace{ID: wsID, Lock: &privilegedUser.ID},
			unprivilegedUser,
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
			&Workspace{ID: wsID, Lock: &privilegedUser.ID},
			privilegedUser,
			LockButton{
				State:   "locked",
				Text:    "Unlock",
				Message: "locked by: janitor",
				Tooltip: "locked by: janitor",
				Action:  "/app/workspaces/ws-123/unlock",
			},
		},
		{
			"user can force unlock lock held by different user",
			&Workspace{ID: wsID, Lock: &privilegedUser.ID},
			privilegedUser2,
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
				service: &fakeUIHelpersService{},
			}
			got, err := helpers.lockButtonHelper(context.Background(), tt.ws, tt.user)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeUIHelpersService struct{}

func (f *fakeUIHelpersService) GetUser(context.Context, user.UserSpec) (*user.User, error) {
	return &user.User{Username: "janitor"}, nil
}
