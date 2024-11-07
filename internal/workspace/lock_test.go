package workspace

import (
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	janitorTestID = resource.NewID(resource.UserKind)
	burglarTestID = resource.NewID(resource.UserKind)
	runTestID1    = resource.NewID(resource.RunKind)
	runTestID2    = resource.NewID(resource.RunKind)
)

func TestWorkspace_Lock(t *testing.T) {
	t.Run("lock an unlocked lock", func(t *testing.T) {
		lock := &lock{}
		err := lock.Enlock(janitorTestID)
		require.NoError(t, err)
		assert.True(t, lock.Locked())
	})
	t.Run("replace run lock with another run lock", func(t *testing.T) {
		lock := &lock{ID: &runTestID1}
		err := lock.Enlock(runTestID2)
		require.NoError(t, err)
		assert.True(t, lock.Locked())
	})
	t.Run("user cannot lock a locked workspace", func(t *testing.T) {
		lock := &lock{ID: &runTestID1}
		err := lock.Enlock(janitorTestID)
		require.Equal(t, ErrWorkspaceAlreadyLocked, err)
	})
}

func TestWorkspace_Unlock(t *testing.T) {
	t.Run("cannot unlock workspace already unlocked", func(t *testing.T) {
		err := (&lock{}).Unlock(janitorTestID, false)
		require.Equal(t, ErrWorkspaceAlreadyUnlocked, err)
	})
	t.Run("user can unlock their own lock", func(t *testing.T) {
		lock := &lock{ID: &janitorTestID}
		err := lock.Unlock(janitorTestID, false)
		require.NoError(t, err)
		assert.False(t, lock.Locked())
	})
	t.Run("user cannot unlock another user's lock", func(t *testing.T) {
		lock := &lock{ID: &janitorTestID}
		err := lock.Unlock(burglarTestID, false)
		require.Equal(t, ErrWorkspaceLockedByDifferentUser, err)
	})
	t.Run("user can unlock a lock by force", func(t *testing.T) {
		lock := &lock{ID: &janitorTestID}
		err := lock.Unlock(burglarTestID, true)
		require.NoError(t, err)
		assert.False(t, lock.Locked())
	})
	t.Run("run can unlock its own lock", func(t *testing.T) {
		runID := resource.ParseID("run-123")
		lock := &lock{ID: &runID}
		err := lock.Unlock(runID, false)
		require.NoError(t, err)
		assert.False(t, lock.Locked())
	})
}
