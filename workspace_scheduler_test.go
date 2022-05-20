package otf

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testActiveRun = Run{
		ID:                   "run-active",
		ConfigurationVersion: &ConfigurationVersion{},
		status:               RunPlanning,
		Workspace:            &Workspace{ID: "ws-123"},
	}

	testPendingSpeculativeRun = Run{
		ID:                   "run-pending-speculative",
		ConfigurationVersion: &ConfigurationVersion{speculative: true},
		status:               RunPending,
		Workspace:            &Workspace{ID: "ws-123"},
	}

	testPendingRun = Run{
		ID:                   "run-pending",
		ConfigurationVersion: &ConfigurationVersion{},
		status:               RunPending,
		Workspace:            &Workspace{ID: "ws-123"},
	}

	testPendingNowActiveRun = Run{
		ID:                   "run-pending",
		ConfigurationVersion: &ConfigurationVersion{},
		status:               RunPlanning,
		Workspace:            &Workspace{ID: "ws-123"},
	}

	testDoneRun = Run{
		ID:                   "run-done",
		ConfigurationVersion: &ConfigurationVersion{},
		status:               RunApplied,
		Workspace:            &Workspace{ID: "ws-123"},
	}

	testWorkspace = Workspace{
		ID: "ws-123",
	}
)

func TestWorkspaceScheduler_New(t *testing.T) {
	tests := []struct {
		name       string
		workspaces []*Workspace
		runs       []*Run
		// wanted workspace queues
		wantQueues map[string][]*Run
		// wanted started runs
		wantStarted []*Run
	}{
		{
			name:       "nothing",
			wantQueues: map[string][]*Run{},
		},
		{
			name:       "one workspace no runs",
			workspaces: []*Workspace{&testWorkspace},
			wantQueues: map[string][]*Run{"ws-123": {}},
		},
		{
			name:        "start pending run",
			workspaces:  []*Workspace{&testWorkspace},
			runs:        []*Run{&testPendingRun},
			wantQueues:  map[string][]*Run{"ws-123": {&testPendingRun}},
			wantStarted: []*Run{&testPendingRun},
		},
		{
			name:        "start speculative run",
			workspaces:  []*Workspace{&testWorkspace},
			runs:        []*Run{&testPendingSpeculativeRun},
			wantQueues:  map[string][]*Run{"ws-123": {}},
			wantStarted: []*Run{&testPendingSpeculativeRun},
		},
		{
			name:       "start no runs",
			workspaces: []*Workspace{&testWorkspace},
			runs: []*Run{
				&testActiveRun,
				&testPendingRun,
			},
			wantQueues: map[string][]*Run{"ws-123": {
				&testActiveRun,
				&testPendingRun,
			}},
		},
		{
			name:       "dequeue done run, start pending run",
			workspaces: []*Workspace{&testWorkspace},
			runs: []*Run{
				&testDoneRun,
				&testPendingRun,
			},
			wantQueues: map[string][]*Run{"ws-123": {
				&testPendingRun,
			}},
			wantStarted: []*Run{&testPendingRun},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := newFakeRunService(tt.runs...)
			scheduler, err := NewWorkspaceScheduler(
				context.Background(),
				&fakeWorkspaceService{workspaces: tt.workspaces},
				rs, nil, logr.Discard())
			require.NoError(t, err)
			assert.Equal(t, tt.wantQueues, scheduler.queues)
			for _, want := range tt.wantStarted {
				assert.Contains(t, rs.started, want.ID)
			}
		})
	}
}

func TestWorkspaceScheduler_Refresh(t *testing.T) {
	tests := []struct {
		name string
		// existing workspace queues before refresh
		existing map[string][]*Run
		// updated run passed to refresh
		updated *Run
		// wanted workspace queues
		wantQueues map[string][]*Run
		// wanted started run
		wantStarted *Run
	}{
		{
			name:        "enqueue and start pending run",
			existing:    map[string][]*Run{"ws-123": {}},
			updated:     &testPendingRun,
			wantQueues:  map[string][]*Run{"ws-123": {&testPendingRun}},
			wantStarted: &testPendingRun,
		},
		{
			name:       "enqueue pending run",
			existing:   map[string][]*Run{"ws-123": {&testActiveRun}},
			updated:    &testPendingRun,
			wantQueues: map[string][]*Run{"ws-123": {&testActiveRun, &testPendingRun}},
		},
		{
			name:        "skip active run and start pending speculative run",
			existing:    map[string][]*Run{"ws-123": {&testActiveRun}},
			updated:     &testPendingSpeculativeRun,
			wantQueues:  map[string][]*Run{"ws-123": {&testActiveRun}},
			wantStarted: &testPendingSpeculativeRun,
		},
		{
			name:       "update pending run to active",
			existing:   map[string][]*Run{"ws-123": {&testPendingRun}},
			updated:    &testPendingNowActiveRun,
			wantQueues: map[string][]*Run{"ws-123": {&testPendingNowActiveRun}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &fakeRunService{}
			scheduler := &WorkspaceScheduler{
				RunService: rs,
				queues:     tt.existing,
			}
			scheduler.refresh(context.Background(), tt.updated)
			assert.Equal(t, tt.wantQueues, scheduler.queues)
			if tt.wantStarted == nil {
				assert.Empty(t, rs.started)
			} else {
				assert.Contains(t, rs.started, tt.wantStarted.ID)
			}
		})
	}
}
