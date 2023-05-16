package workspace

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_Lock(t *testing.T) {
	t.Run("lock an unlocked workspace", func(t *testing.T) {
		ws := &Workspace{}
		err := ws.Enlock("janitor", UserLock)
		require.NoError(t, err)
		assert.True(t, ws.Locked())
	})
	t.Run("replace run lock with another run lock", func(t *testing.T) {
		ws := &Workspace{Lock: &Lock{id: "run-123", LockKind: RunLock}}
		err := ws.Enlock("run-456", RunLock)
		require.NoError(t, err)
		assert.True(t, ws.Locked())
	})
	t.Run("user cannot lock a locked workspace", func(t *testing.T) {
		ws := &Workspace{Lock: &Lock{id: "run-123", LockKind: RunLock}}
		err := ws.Enlock("janitor", UserLock)
		require.Equal(t, internal.ErrWorkspaceAlreadyLocked, err)
	})
}

func TestWorkspace_Unlock(t *testing.T) {
	t.Run("cannot unlock workspace already unlocked", func(t *testing.T) {
		err := (&Workspace{}).Unlock("janitor", UserLock, false)
		require.Equal(t, internal.ErrWorkspaceAlreadyUnlocked, err)
	})
	t.Run("user can unlock their own lock", func(t *testing.T) {
		ws := &Workspace{Lock: &Lock{id: "janitor", LockKind: UserLock}}
		err := ws.Unlock("janitor", UserLock, false)
		require.NoError(t, err)
		assert.False(t, ws.Locked())
	})
	t.Run("user cannot unlock another user's lock", func(t *testing.T) {
		ws := &Workspace{Lock: &Lock{id: "janitor", LockKind: UserLock}}
		err := ws.Unlock("burglar", UserLock, false)
		require.Equal(t, internal.ErrWorkspaceLockedByDifferentUser, err)
	})
	t.Run("user can unlock a lock by force", func(t *testing.T) {
		ws := &Workspace{Lock: &Lock{id: "janitor", LockKind: UserLock}}
		err := ws.Unlock("headmaster", UserLock, true)
		require.NoError(t, err)
		assert.False(t, ws.Locked())
	})
	t.Run("run can unlock its own lock", func(t *testing.T) {
		ws := &Workspace{Lock: &Lock{id: "run-123", LockKind: RunLock}}
		err := ws.Unlock("run-123", RunLock, false)
		require.NoError(t, err)
		assert.False(t, ws.Locked())
	})
}

func TestWorkspace_LockButtonHelper(t *testing.T) {
	tests := []struct {
		name    string
		ws      *Workspace
		subject *fakeSubject
		want    LockButton
	}{
		{
			"unlocked state",
			&Workspace{ID: "ws-123"},
			&fakeSubject{canLock: true},
			LockButton{
				State:  "unlocked",
				Text:   "Lock",
				Action: "/app/workspaces/ws-123/lock",
			},
		},
		{
			"insufficient permissions to lock",
			&Workspace{ID: "ws-123"},
			&fakeSubject{},
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
			&Workspace{Lock: &Lock{id: "janitor", LockKind: UserLock}},
			&fakeSubject{},
			LockButton{
				State:    "locked",
				Text:     "Unlock",
				Tooltip:  "insufficient permissions",
				Action:   "/app/workspaces//unlock",
				Disabled: true,
			},
		},
		{
			"user can unlock their own lock",
			&Workspace{Lock: &Lock{id: "janitor", LockKind: UserLock}},
			&fakeSubject{id: "janitor", canUnlock: true},
			LockButton{
				State:   "locked",
				Text:    "Unlock",
				Tooltip: "locked by user: janitor",
				Action:  "/app/workspaces//unlock",
			},
		},
		{
			"can unlock lock held by a different user",
			&Workspace{Lock: &Lock{id: "janitor", LockKind: UserLock}},
			&fakeSubject{id: "burglar", canUnlock: true},
			LockButton{
				State:    "locked",
				Text:     "Unlock",
				Action:   "/app/workspaces//unlock",
				Tooltip:  "locked by user: janitor",
				Disabled: true,
			},
		},
		{
			"user can force unlock",
			&Workspace{Lock: &Lock{id: "janitor", LockKind: UserLock}},
			&fakeSubject{id: "headmaster", canUnlock: true, canForceUnlock: true},
			LockButton{
				State:   "locked",
				Text:    "Force unlock",
				Action:  "/app/workspaces//force-unlock",
				Tooltip: "locked by user: janitor",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lockButtonHelper(tt.ws, internal.WorkspacePolicy{}, tt.subject)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeSubject struct {
	id                                 string
	canUnlock, canForceUnlock, canLock bool

	internal.Subject
}

func (f *fakeSubject) String() string { return f.id }

func (f *fakeSubject) CanAccessWorkspace(action rbac.Action, _ internal.WorkspacePolicy) bool {
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
