package inmem

import (
	"testing"

	"github.com/go-logr/logr"
	tfe "github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScheduler(t *testing.T) {
	scheduler, err := NewScheduler(
		&mock.WorkspaceService{
			ListWorkspaceFn: func(opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error) {
				return &ots.WorkspaceList{
					Items: []*ots.Workspace{
						{
							ID: "ws-123",
						},
					},
				}, nil
			},
		},
		&mock.RunService{
			// Mock results depending on whether active or pending runs are
			// requested.
			ListFn: func(opts ots.RunListOptions) (*ots.RunList, error) {
				for _, status := range opts.Statuses {
					switch status {
					case tfe.RunPlanning:
						return &ots.RunList{
							Items: []*ots.Run{
								{
									ID:                   "run-active",
									ConfigurationVersion: &ots.ConfigurationVersion{},
									Status:               tfe.RunPlanning,
								},
								{
									ID:                   "run-speculative",
									ConfigurationVersion: &ots.ConfigurationVersion{Speculative: true},
									Status:               tfe.RunPlanning,
								},
							},
						}, nil
					case tfe.RunPending:
						return &ots.RunList{
							Items: []*ots.Run{
								{
									ID:                   "run-pending",
									ConfigurationVersion: &ots.ConfigurationVersion{},
									Status:               tfe.RunPending,
								},
							},
						}, nil
					}
				}
				return nil, nil
			},
		},
		nil,
		logr.Discard(),
	)
	require.NoError(t, err)
	require.NotNil(t, scheduler)

	// Expecting
	// queues=map[ws-123]Queue{active:run-active,pending:[run-pending]}
	if assert.Contains(t, scheduler.Queues, "ws-123") {
		if assert.NotNil(t, scheduler.Queues["ws-123"].(*ots.WorkspaceQueue).Active) {
			assert.Equal(t, "run-active", scheduler.Queues["ws-123"].(*ots.WorkspaceQueue).Active.ID)
		}
		if assert.Equal(t, 1, len(scheduler.Queues["ws-123"].(*ots.WorkspaceQueue).Pending)) {
			assert.Equal(t, "run-pending", scheduler.Queues["ws-123"].(*ots.WorkspaceQueue).Pending[0].ID)
		}
	}
}

func TestScheduler_AddWorkspace(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: make(map[string]ots.Queue),
	}

	scheduler.handleEvent(ots.Event{
		Type:    ots.WorkspaceCreated,
		Payload: &ots.Workspace{ID: "ws-123"},
	})

	assert.Contains(t, scheduler.Queues, "ws-123")
}

func TestScheduler_RemoveWorkspace(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]ots.Queue{
			"ws-123": &mock.Queue{},
		},
	}

	scheduler.handleEvent(ots.Event{
		Type:    ots.WorkspaceDeleted,
		Payload: &ots.Workspace{ID: "ws-123"},
	})

	assert.NotContains(t, scheduler.Queues, "ws-123")
}

func TestScheduler_AddRun(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]ots.Queue{
			"ws-123": &mock.Queue{},
		},
	}

	scheduler.handleEvent(ots.Event{
		Type: ots.RunCreated,
		Payload: &ots.Run{
			ID: "ws-123",
			Workspace: &ots.Workspace{
				ID:           "ws-123",
				Organization: &ots.Organization{ID: "org-123"},
			},
		},
	})

	assert.Equal(t, 1, len(scheduler.Queues["ws-123"].(*mock.Queue).Runs))
}

func TestScheduler_RemoveRun(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]ots.Queue{
			"ws-123": &mock.Queue{
				Runs: []*ots.Run{
					{
						ID: "run-123",
					},
				},
			},
		},
	}
	require.NotNil(t, scheduler)

	scheduler.handleEvent(ots.Event{
		Type: ots.RunCompleted,
		Payload: &ots.Run{
			ID: "run-123",
			Workspace: &ots.Workspace{
				ID:           "ws-123",
				Organization: &ots.Organization{ID: "org-123"},
			},
		},
	})

	assert.Equal(t, 0, len(scheduler.Queues["ws-123"].(*mock.Queue).Runs))
}
