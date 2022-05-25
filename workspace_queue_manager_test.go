package otf

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceQueueManager_Seed(t *testing.T) {
	ws1speculative1 := NewTestRun("ws1-spec-1", TestRunWorkspaceID("ws1"), TestRunSpeculative())
	ws1pending1 := NewTestRun("ws1-pending-1", TestRunWorkspaceID("ws1"))
	ws1pending2 := NewTestRun("ws1-pending-2", TestRunWorkspaceID("ws1"))
	ws1pending3 := NewTestRun("ws1-pending-3", TestRunWorkspaceID("ws1"))

	ws2planQueued1 := NewTestRun("ws1-plan-queued1", TestRunWorkspaceID("ws2"), TestRunStatus(RunPlanQueued))
	ws2pending1 := NewTestRun("ws2-pending-1", TestRunWorkspaceID("ws2"))
	ws2speculative1 := NewTestRun("ws2-spec-1", TestRunWorkspaceID("ws2"), TestRunSpeculative())
	ws2pending2 := NewTestRun("ws2-pending-2", TestRunWorkspaceID("ws2"))

	// order matters, with oldest run first
	rs := newFakeRunService(
		ws1speculative1, ws1pending1, ws1pending2, ws1pending3,
		ws2planQueued1, ws2pending1, ws2speculative1, ws2pending2)

	mgr := &workspaceQueueManager{
		RunService: rs,
		Logger:     logr.Discard(),
		ctx:        context.Background(),
		queues:     make(map[string]workspaceQueue),
	}
	err := mgr.seed()
	require.NoError(t, err)

	assert.Equal(t, map[string]workspaceQueue{
		"ws1": {ws1pending1, ws1pending2, ws1pending3},
		"ws2": {ws2planQueued1, ws2pending1, ws2pending2},
	}, mgr.queues)
}

func TestWorkspaceQueueManager_Refresh(t *testing.T) {
	ws1speculative1 := NewTestRun("ws1-spec-1", TestRunWorkspaceID("ws1"), TestRunSpeculative())
	ws1planQueued1 := NewTestRun("ws1-plan-queued1", TestRunWorkspaceID("ws1"), TestRunStatus(RunPlanQueued))
	ws1pending1 := NewTestRun("ws1-pending-1", TestRunWorkspaceID("ws1"))
	ws1done1 := NewTestRun("ws1-done-1", TestRunWorkspaceID("ws1"), TestRunStatus(RunPlannedAndFinished))

	tests := []struct {
		name string
		// updated run passed to refresh
		run *Run
		// existing queues before refresh
		existing map[string]workspaceQueue
		// wanted workspace queues
		want map[string]workspaceQueue
	}{
		{
			name:     "enqueue pending to empty queue",
			run:      ws1pending1,
			existing: map[string]workspaceQueue{},
			want:     map[string]workspaceQueue{"ws1": {ws1pending1}},
		},
		{
			name:     "enqueue pending to non-empty queue",
			run:      ws1pending1,
			existing: map[string]workspaceQueue{"ws1": {ws1planQueued1}},
			want:     map[string]workspaceQueue{"ws1": {ws1planQueued1, ws1pending1}},
		},
		{
			name:     "dont enqueue speculative",
			run:      ws1speculative1,
			existing: map[string]workspaceQueue{"ws1": {}},
			want:     map[string]workspaceQueue{"ws1": {}},
		},
		{
			name:     "remove done",
			run:      ws1done1,
			existing: map[string]workspaceQueue{"ws1": {ws1done1, ws1pending1}},
			want:     map[string]workspaceQueue{"ws1": {ws1pending1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := newFakeRunService(append(tt.existing["ws1"], tt.run)...)
			mgr := &workspaceQueueManager{
				RunService: rs,
				Logger:     logr.Discard(),
				ctx:        context.Background(),
				queues:     tt.existing,
			}
			err := mgr.refresh(tt.run)
			require.NoError(t, err)
			assert.Equal(t, tt.want, mgr.queues)
		})
	}
}
