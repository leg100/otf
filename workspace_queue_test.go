package ots

import (
	"testing"

	tfe "github.com/leg100/go-tfe"
	"github.com/stretchr/testify/assert"
)

type mockRunStatusUpdater struct{}

func (u *mockRunStatusUpdater) UpdateStatus(string, tfe.RunStatus) (*Run, error) { return nil, nil }

func TestWorkspaceQueue_AddRun(t *testing.T) {
	tests := []struct {
		name    string
		active  *Run
		pending []*Run
		run     *Run
		want    func(*testing.T, *Run, []*Run)
	}{
		{
			name: "no existing runs",
			run:  &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{}},
			want: func(t *testing.T, active *Run, pending []*Run) {
				if assert.NotNil(t, active) {
					assert.Equal(t, "run-123", active.ID)
				}
				assert.Equal(t, 0, len(pending))
			},
		},
		{
			name:   "existing active run",
			active: &Run{ID: "run-active"},
			run:    &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{}},
			want: func(t *testing.T, active *Run, pending []*Run) {
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
			active:  &Run{ID: "run-active"},
			pending: []*Run{{ID: "run-pending"}},
			run:     &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{}},
			want: func(t *testing.T, active *Run, pending []*Run) {
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
			run:  &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{Speculative: true}},
			want: func(t *testing.T, active *Run, pending []*Run) {
				assert.Nil(t, active)
				assert.Equal(t, 0, len(pending))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &WorkspaceQueue{
				RunStatusUpdater: &mockRunStatusUpdater{},
				Active:           tt.active,
				Pending:          tt.pending,
			}

			q.Add(tt.run)

			tt.want(t, q.Active, q.Pending)
		})
	}
}

func TestWorkspaceQueue_RemoveRun(t *testing.T) {
	tests := []struct {
		name    string
		active  *Run
		pending []*Run
		run     *Run
		want    func(*testing.T, *Run, []*Run)
	}{
		{
			name:   "remove active run",
			active: &Run{ID: "run-123"},
			run:    &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{}},
			want: func(t *testing.T, active *Run, pending []*Run) {
				assert.Nil(t, active)
			},
		},
		{
			name:   "remove active run, pending run takes its place",
			active: &Run{ID: "run-123"},
			pending: []*Run{
				{ID: "run-pending-0"},
				{ID: "run-pending-1"},
			},
			run: &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{}},
			want: func(t *testing.T, active *Run, pending []*Run) {
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
			active:  &Run{ID: "run-active"},
			pending: []*Run{{ID: "run-123"}},
			run:     &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{}},
			want: func(t *testing.T, active *Run, pending []*Run) {
				assert.Equal(t, 0, len(pending))
			},
		},
		{
			name:   "remove pending run amongst other pending runs",
			active: &Run{ID: "run-active"},
			pending: []*Run{
				{ID: "run-pending-0"},
				{ID: "run-123"},
				{ID: "run-pending-1"},
			},
			run: &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{}},
			want: func(t *testing.T, active *Run, pending []*Run) {
				if assert.Equal(t, 2, len(pending)) {
					assert.Equal(t, "run-pending-0", pending[0].ID)
					assert.Equal(t, "run-pending-1", pending[1].ID)
				}
			},
		},
		{
			name:    "remove speculative run",
			active:  &Run{ID: "run-active"},
			pending: []*Run{{ID: "run-pending"}},
			run:     &Run{ID: "run-123", ConfigurationVersion: &ConfigurationVersion{Speculative: true}},
			want: func(t *testing.T, active *Run, pending []*Run) {
				assert.Equal(t, "run-active", active.ID)
				if assert.Equal(t, 1, len(pending)) {
					assert.Equal(t, "run-pending", pending[0].ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &WorkspaceQueue{
				RunStatusUpdater: &mockRunStatusUpdater{},
				Active:           tt.active,
				Pending:          tt.pending,
			}

			q.Remove(tt.run)

			tt.want(t, q.Active, q.Pending)
		})
	}
}
