package inmem

import (
	"testing"

	"github.com/go-logr/logr"
	tfe "github.com/leg100/go-tfe"
	"github.com/leg100/otf"
	"github.com/leg100/otf/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScheduler(t *testing.T) {
	scheduler, err := NewScheduler(
		&mock.WorkspaceService{
			ListWorkspaceFn: func(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
				return &otf.WorkspaceList{
					Items: []*otf.Workspace{
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
			ListFn: func(opts otf.RunListOptions) (*otf.RunList, error) {
				for _, status := range opts.Statuses {
					switch status {
					case tfe.RunPlanning:
						return &otf.RunList{
							Items: []*otf.Run{
								{
									ID:                   "run-active",
									ConfigurationVersion: &otf.ConfigurationVersion{},
									Status:               tfe.RunPlanning,
								},
								{
									ID:                   "run-speculative",
									ConfigurationVersion: &otf.ConfigurationVersion{Speculative: true},
									Status:               tfe.RunPlanning,
								},
							},
						}, nil
					case tfe.RunPending:
						return &otf.RunList{
							Items: []*otf.Run{
								{
									ID:                   "run-pending",
									ConfigurationVersion: &otf.ConfigurationVersion{},
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
		if assert.NotNil(t, scheduler.Queues["ws-123"].(*otf.WorkspaceQueue).Active) {
			assert.Equal(t, "run-active", scheduler.Queues["ws-123"].(*otf.WorkspaceQueue).Active.ID)
		}
		if assert.Equal(t, 1, len(scheduler.Queues["ws-123"].(*otf.WorkspaceQueue).Pending)) {
			assert.Equal(t, "run-pending", scheduler.Queues["ws-123"].(*otf.WorkspaceQueue).Pending[0].ID)
		}
	}
}

func TestScheduler_AddWorkspace(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: make(map[string]otf.Queue),
	}

	scheduler.handleEvent(otf.Event{
		Type:    otf.WorkspaceCreated,
		Payload: &otf.Workspace{ID: "ws-123"},
	})

	assert.Contains(t, scheduler.Queues, "ws-123")
}

func TestScheduler_RemoveWorkspace(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]otf.Queue{
			"ws-123": &mock.Queue{},
		},
	}

	scheduler.handleEvent(otf.Event{
		Type:    otf.WorkspaceDeleted,
		Payload: &otf.Workspace{ID: "ws-123"},
	})

	assert.NotContains(t, scheduler.Queues, "ws-123")
}

func TestScheduler_AddRun(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]otf.Queue{
			"ws-123": &mock.Queue{},
		},
	}

	scheduler.handleEvent(otf.Event{
		Type: otf.RunCreated,
		Payload: &otf.Run{
			ID: "ws-123",
			Workspace: &otf.Workspace{
				ID:           "ws-123",
				Organization: &otf.Organization{ID: "org-123"},
			},
		},
	})

	assert.Equal(t, 1, len(scheduler.Queues["ws-123"].(*mock.Queue).Runs))
}

func TestScheduler_RemoveRun(t *testing.T) {
	scheduler := &Scheduler{
		Logger: logr.Discard(),
		Queues: map[string]otf.Queue{
			"ws-123": &mock.Queue{
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
		Type: otf.RunCompleted,
		Payload: &otf.Run{
			ID: "run-123",
			Workspace: &otf.Workspace{
				ID:           "ws-123",
				Organization: &otf.Organization{ID: "org-123"},
			},
		},
	})

	assert.Equal(t, 0, len(scheduler.Queues["ws-123"].(*mock.Queue).Runs))
}
