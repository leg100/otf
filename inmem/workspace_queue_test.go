package inmem

import (
	"testing"

	"github.com/go-logr/logr"
	tfe "github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceQueue_AddRun(t *testing.T) {
	tests := []struct {
		name    string
		active  *ots.Run
		pending []*ots.Run
		run     *ots.Run
		want    func(*testing.T, *ots.Run, []*ots.Run)
	}{
		{
			name: "no existing runs",
			run:  &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				if assert.NotNil(t, active) {
					assert.Equal(t, "run-123", active.ID)
				}
				assert.Equal(t, 0, len(pending))
			},
		},
		{
			name:   "existing active run",
			active: &ots.Run{ID: "run-active"},
			run:    &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				if assert.NotNil(t, active) {
					assert.Equal(t, "run-active", active.ID)
				}
				if assert.Equal(t, 1, len(pending)) {
					assert.Equal(t, "run-123", pending[0].ID)
				}
			},
		},
		{
			name:    "existing active run and pending run",
			active:  &ots.Run{ID: "run-active"},
			pending: []*ots.Run{{ID: "run-pending"}},
			run:     &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				if assert.NotNil(t, active) {
					assert.Equal(t, "run-active", active.ID)
				}
				if assert.Equal(t, 2, len(pending)) {
					assert.Equal(t, "run-pending", pending[0].ID)
					assert.Equal(t, "run-123", pending[1].ID)
				}
			},
		},
		{
			name: "add speculative run",
			run:  &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{Speculative: true}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				assert.Nil(t, active)
				assert.Equal(t, 0, len(pending))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewWorkspaceQueue(
				&mock.RunService{
					UpdateStatusFn: func(id string, status tfe.RunStatus) (*ots.Run, error) {
						return nil, nil
					},
				},
				logr.Discard(),
				"ws-123",
				WithActive(tt.active),
				WithPending(tt.pending),
			)

			q.addRun(tt.run)

			tt.want(t, q.active, q.pending)
		})
	}
}

func TestWorkspaceQueue_RemoveRun(t *testing.T) {
	tests := []struct {
		name    string
		active  *ots.Run
		pending []*ots.Run
		run     *ots.Run
		want    func(*testing.T, *ots.Run, []*ots.Run)
	}{
		{
			name:   "remove active run",
			active: &ots.Run{ID: "run-123"},
			run:    &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				assert.Nil(t, active)
			},
		},
		{
			name:   "remove active run, pending run takes its place",
			active: &ots.Run{ID: "run-123"},
			pending: []*ots.Run{
				{ID: "run-pending-0"},
				{ID: "run-pending-1"},
			},
			run: &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				if assert.NotNil(t, active) {
					assert.Equal(t, "run-pending-0", active.ID)
				}
				if assert.Equal(t, 1, len(pending)) {
					assert.Equal(t, "run-pending-1", pending[0].ID)
				}
			},
		},
		{
			name:    "remove only pending run",
			active:  &ots.Run{ID: "run-active"},
			pending: []*ots.Run{{ID: "run-123"}},
			run:     &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				assert.Equal(t, 0, len(pending))
			},
		},
		{
			name:   "remove pending run amongst other pending runs",
			active: &ots.Run{ID: "run-active"},
			pending: []*ots.Run{
				{ID: "run-pending-0"},
				{ID: "run-123"},
				{ID: "run-pending-1"},
			},
			run: &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				if assert.Equal(t, 2, len(pending)) {
					assert.Equal(t, "run-pending-0", pending[0].ID)
					assert.Equal(t, "run-pending-1", pending[1].ID)
				}
			},
		},
		{
			name:    "remove speculative run",
			active:  &ots.Run{ID: "run-active"},
			pending: []*ots.Run{{ID: "run-pending"}},
			run:     &ots.Run{ID: "run-123", ConfigurationVersion: &ots.ConfigurationVersion{Speculative: true}},
			want: func(t *testing.T, active *ots.Run, pending []*ots.Run) {
				assert.Equal(t, "run-active", active.ID)
				if assert.Equal(t, 1, len(pending)) {
					assert.Equal(t, "run-pending", pending[0].ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewWorkspaceQueue(
				&mock.RunService{
					UpdateStatusFn: func(id string, status tfe.RunStatus) (*ots.Run, error) {
						return nil, nil
					},
				},
				logr.Discard(),
				"ws-123",
				WithActive(tt.active),
				WithPending(tt.pending),
			)

			q.removeRun(tt.run)

			tt.want(t, q.active, q.pending)
		})
	}
}
