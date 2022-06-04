package otf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunFactory(t *testing.T) {
	tests := []struct {
		name string
		opts RunCreateOptions
		want func(*testing.T, *Run, error)
		ws   Workspace
		cv   ConfigurationVersion
	}{
		{
			name: "defaults",
			ws:   Workspace{id: "ws-123"},
			cv:   ConfigurationVersion{id: "cv-123"},
			opts: RunCreateOptions{},
			want: func(t *testing.T, run *Run, err error) {
				assert.Equal(t, RunPending, run.status)
				assert.NotZero(t, run.createdAt)
				assert.False(t, run.speculative)
				assert.True(t, run.refresh)
				assert.False(t, run.autoApply)
			},
		},
		{
			name: "speculative",
			ws:   Workspace{id: "ws-123"},
			cv:   ConfigurationVersion{id: "cv-123", speculative: true},
			opts: RunCreateOptions{},
			want: func(t *testing.T, run *Run, err error) {
				assert.True(t, run.speculative)
				assert.Equal(t, RunPlanQueued, run.status)
			},
		},
		{
			name: "auto-apply",
			ws:   Workspace{id: "ws-123", autoApply: true},
			cv:   ConfigurationVersion{id: "cv-123"},
			opts: RunCreateOptions{},
			want: func(t *testing.T, run *Run, err error) {
				assert.True(t, run.autoApply)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := RunFactory{
				WorkspaceService:            &fakeRunFactoryWorkspaceService{ws: &tt.ws},
				ConfigurationVersionService: &fakeRunFactoryConfigurationVersionService{cv: &tt.cv},
			}
			run, err := f.New(context.Background(), tt.ws.SpecID(), tt.opts)
			tt.want(t, run, err)
		})
	}
}
