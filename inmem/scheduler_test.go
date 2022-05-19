package inmem

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testActiveRun = otf.Run{
		ID:                   "run-active",
		ConfigurationVersion: &otf.ConfigurationVersion{},
		Status:               otf.RunPlanning,
		Workspace:            &otf.Workspace{ID: "ws-123"},
	}

	testSpeculativeRun = otf.Run{
		ID:                   "run-speculative",
		ConfigurationVersion: &otf.ConfigurationVersion{Speculative: true},
		Status:               otf.RunPlanning,
		Workspace:            &otf.Workspace{ID: "ws-123"},
	}

	testPendingRun = otf.Run{
		ID:                   "run-pending",
		ConfigurationVersion: &otf.ConfigurationVersion{},
		Status:               otf.RunPending,
		Workspace:            &otf.Workspace{ID: "ws-123"},
	}

	testWorkspace = otf.Workspace{
		ID: "ws-123",
	}
)

func TestNewScheduler_PopulateQueues(t *testing.T) {
	tests := []struct {
		name       string
		runs       []*otf.Run
		workspaces []*otf.Workspace
		want       map[string]*otf.WorkspaceQueue
	}{
		{
			name: "nothing",
			runs: []*otf.Run{
				&testActiveRun,
				&testSpeculativeRun,
				&testPendingRun,
			},
			workspaces: []*otf.Workspace{
				&testWorkspace,
			},
			want: map[string]*otf.WorkspaceQueue{
				"ws-123": &otf.WorkspaceQueue{
					Active:  &testActiveRun,
					Pending: []*otf.Run{&testPendingRun},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewScheduler(
				&fakeWorkspaceService{workspaces: tt.workspaces},
				&fakeRunService{runs: tt.runs},
				nil, logr.Discard())
			require.NoError(t, err)
			require.NotNil(t, scheduler)

			for wid, q := range tt.want {
				if assert.Contains(t, scheduler.Queues, wid) {
					assert.Equal(t, q.Active, scheduler.Queues[wid].(*otf.WorkspaceQueue).Active)
					assert.Equal(t, q.Pending, scheduler.Queues[wid].(*otf.WorkspaceQueue).Pending)
				}
			}
		})
	}
}

func TestScheduler_AddWorkspace(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: make(map[string]otf.Queue),
	}

	scheduler.handleEvent(otf.Event{
		Type:    otf.EventWorkspaceCreated,
		Payload: &otf.Workspace{ID: "ws-123"},
	})

	assert.Contains(t, scheduler.Queues, "ws-123")
}

func TestScheduler_RemoveWorkspace(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]otf.Queue{
			"ws-123": &fakeQueue{},
		},
	}

	scheduler.handleEvent(otf.Event{
		Type:    otf.EventWorkspaceDeleted,
		Payload: &otf.Workspace{ID: "ws-123"},
	})

	assert.NotContains(t, scheduler.Queues, "ws-123")
}

func TestScheduler_AddRun(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]otf.Queue{
			"ws-123": &fakeQueue{},
		},
	}

	scheduler.handleEvent(otf.Event{
		Type: otf.EventRunCreated,
		Payload: &otf.Run{
			ID: "ws-123",
			Workspace: &otf.Workspace{
				ID:           "ws-123",
				Organization: &otf.Organization{ID: "org-123"},
			},
		},
	})

	assert.Equal(t, 1, len(scheduler.Queues["ws-123"].(*fakeQueue).Runs))
}

func TestScheduler_RemoveRun(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]otf.Queue{
			"ws-123": &fakeQueue{
				Runs: []*otf.Run{
					{
						ID: "run-123",
					},
				},
			},
		},
	}
	require.NotNil(t, scheduler)

	scheduler.handleEvent(otf.Event{
		Type: otf.EventRunCompleted,
		Payload: &otf.Run{
			ID: "run-123",
			Workspace: &otf.Workspace{
				ID:           "ws-123",
				Organization: &otf.Organization{ID: "org-123"},
			},
		},
	})

	assert.Equal(t, 0, len(scheduler.Queues["ws-123"].(*fakeQueue).Runs))
}
