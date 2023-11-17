package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_IsAllowed(t *testing.T) {
	assert.True(t, WorkspacePlanRole.IsAllowed(CreateRunAction))
	assert.True(t, WorkspacePlanRole.IsAllowed(ListRunsAction))
	assert.True(t, WorkspacePlanRole.IsAllowed(TailLogsAction))

	assert.True(t, WorkspaceWriteRole.IsAllowed(ApplyRunAction))
	assert.True(t, WorkspaceWriteRole.IsAllowed(CancelRunAction))
	assert.True(t, WorkspaceWriteRole.IsAllowed(CreateRunAction))
	assert.True(t, WorkspaceWriteRole.IsAllowed(ListRunsAction))
	assert.True(t, WorkspaceWriteRole.IsAllowed(TailLogsAction))

	assert.True(t, WorkspaceAdminRole.IsAllowed(SetWorkspacePermissionAction))
	assert.True(t, WorkspaceAdminRole.IsAllowed(ApplyRunAction))
	assert.True(t, WorkspaceWriteRole.IsAllowed(CancelRunAction))
	assert.True(t, WorkspaceAdminRole.IsAllowed(CreateRunAction))
	assert.True(t, WorkspaceAdminRole.IsAllowed(ListRunsAction))
	assert.True(t, WorkspaceAdminRole.IsAllowed(TailLogsAction))

	assert.True(t, WorkspaceManagerRole.IsAllowed(CreateWorkspaceAction))
	assert.True(t, WorkspaceManagerRole.IsAllowed(SetWorkspacePermissionAction))
	assert.True(t, WorkspaceManagerRole.IsAllowed(ApplyRunAction))
	assert.True(t, WorkspaceWriteRole.IsAllowed(CancelRunAction))
	assert.True(t, WorkspaceManagerRole.IsAllowed(CreateRunAction))
	assert.True(t, WorkspaceManagerRole.IsAllowed(ListRunsAction))
	assert.True(t, WorkspaceManagerRole.IsAllowed(TailLogsAction))
}
